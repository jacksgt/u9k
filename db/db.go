package db

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"u9k/config"
	"u9k/models"
	"u9k/types"
)

// shared connection pool
var pool *pgxpool.Pool

func InitDBConnection(forceVersion int) {
	rand.Seed(time.Now().UnixNano())

	// set up connection configuration
	conf, err := pgxpool.ParseConfig(config.DbConnUrl)
	if err != nil {
		log.Fatal("error configuring the database: ", err)
	}

	// connect to the database

	// use "pgxpool.Connect()" instead of "pgx.Connect()" because the vanilla driver is not safe
	// for concurrent connections (unlike the other Golang SQL drivers)
	// https://github.com/jackc/pgx/wiki/Getting-started-with-pgx
	pool, err = pgxpool.ConnectConfig(context.Background(), conf)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	//defer pool.Close(context.Background())

	// get a single connection from the pool
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Failed to communcate with database: %s\n", err)
	}
	defer conn.Release()

	// check if connection is working
	err = conn.Conn().Ping(context.Background())
	if err != nil {
		log.Fatalf("Failed to ping database: %s\n", config.DbConnUrl)
	}

	log.Printf("Connected to database %s\n", config.DbConnUrl)

	// run migrations
	//	err = applyMigrations(config.DbConnUrl, forceVersion)
	err = applyMigrations(conn.Conn(), forceVersion)
	if err != nil {
		log.Fatalf("Failed to apply database migrations: %s\n", err)
	}
}

func StoreLink(link *models.Link) error {
	var err error
	if link.Id != "" {
		err = pool.QueryRow(context.Background(),
			"INSERT INTO links (id, url) VALUES ($1, $2) RETURNING id, create_ts",
			link.Id,
			link.Url,
		).Scan(&link.Id, &link.CreateTimestamp)
	} else {
		err = pool.QueryRow(context.Background(),
			"INSERT INTO links (url) VALUES ($1) RETURNING id, create_ts",
			link.Url,
		).Scan(&link.Id, &link.CreateTimestamp)
	}

	// TODO: ideally this should differentiate between generic errors
	// and duplicate key errors
	if err != nil {
		log.Printf("Failed to insert link: %s\n", err)
		return err
	}

	return nil
}

func GetLink(id string) *models.Link {
	link := new(models.Link)
	err := pool.QueryRow(context.Background(),
		"SELECT id, url, create_ts, counter FROM links WHERE id = $1",
		id,
	).Scan(&link.Id, &link.Url, &link.CreateTimestamp, &link.Counter)
	if err != nil {
		log.Printf("Failed to retrieve link %s: %s\n", id, err)
		return nil
	}

	return link
}

func StoreFile(file *models.File) error {
	err := pool.QueryRow(context.Background(),
		"INSERT INTO files (filename, filetype, expire) VALUES ($1, $2, $3) RETURNING id, create_ts",
		file.Name,
		file.Type,
		time.Duration(file.Expire), // cast to time.Duration so pgx knows how to treat the type
	).Scan(&file.Id, &file.CreateTimestamp)
	if err != nil {
		log.Printf("Failed to insert file: %s\n", err)
		return err
	}

	return nil
}

func GetFile(id string) *models.File {
	var expire time.Duration
	file := new(models.File)
	err := pool.QueryRow(context.Background(),
		"SELECT id, filename, filetype, create_ts, counter, expire FROM files WHERE id = $1",
		id,
	).Scan(&file.Id, &file.Name, &file.Type, &file.CreateTimestamp, &file.Counter, &expire)
	if err != nil {
		log.Printf("Failed to retrieve file %s from DB: %s\n", id, err)
		return nil
	}
	file.Expire = types.Duration(expire)

	return file
}

func DeleteFile(id string) error {
	err := pool.QueryRow(context.Background(),
		"DELETE FROM files WHERE id = $1 LIMIT 1 RETURNING id",
		id,
	).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func GetExpiredFiles() ([]models.File, error) {
	files := make([]models.File, 0)
	rows, err := pool.Query(context.Background(),
		"SELECT id, filename, filetype, create_ts, counter, expire FROM files WHERE create_ts + expire < NOW()",
	)
	if err != nil {
		return files, err
	}
	defer rows.Close()

	for rows.Next() {
		var expire time.Duration
		var file models.File
		err := rows.Scan(&file.Id, &file.Name, &file.Type, &file.CreateTimestamp, &file.Counter, &expire)
		if err != nil {
			if err == pgx.ErrNoRows {
				return files, nil
			}
			return files, err
		}
		file.Expire = types.Duration(expire)
		files = append(files, file)
	}

	return files, nil
}

func IncrementLinkCounter(id string) int64 {
	var counter int64
	err := pool.QueryRow(context.Background(),
		"UPDATE links SET counter = counter + 1 WHERE id = $1 RETURNING counter",
		id,
	).Scan(&counter)

	if err != nil {
		log.Printf("Failed to increment link counter: %s\n", err)
		return counter
	}

	return counter
}

func IncrementCounter(typ string, id string) int64 {
	var query string
	if typ == "link" {
		query = "UPDATE links SET counter = counter + 1 WHERE id = $1 RETURNING counter"
	} else if typ == "file" {
		query = "UPDATE files SET counter = counter + 1 WHERE id = $1 RETURNING counter"
	} else {
		return 0
	}

	var counter int64
	err := pool.QueryRow(context.Background(),
		query,
		id,
	).Scan(&counter)

	if err != nil {
		log.Printf("Failed to increment %s counter %s: %s\n", typ, id, err)
		return 0
	}

	return counter
}

// from https://stackoverflow.com/a/31832326
const letterBytes = "abcdefghijklmnopqrstuvwxyz23456789" // omit 1 and 0 for readability
func randStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

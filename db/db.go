package db

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"

	"u9k/config"
	"u9k/types"
)

// shared connection handler
var conn *pgx.Conn

func InitDBConnection(forceVersion int) {
	rand.Seed(time.Now().UnixNano())

	// set up connection configuration
	conf, err := pgx.ParseConfig(config.DbConnUrl)
	if err != nil {
		log.Fatal("error configuring the database: ", err)
	}

	// connect to the database
	conn, err = pgx.ConnectConfig(context.Background(), conf)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	//defer conn.Close(context.Background())

	// check if connection is working
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to communcate with database: ", config.DbConnUrl)
	}

	log.Printf("Connected to database %s\n", config.DbConnUrl)

	// run migrations
	err = applyMigrations(config.DbConnUrl, forceVersion)
	if err != nil {
		log.Fatalf("Failed to apply database migrations: %s\n", err)
	}
}

func StoreLink(link *types.Link) error {
	var err error
	if link.Id != "" {
		err = conn.QueryRow(context.Background(),
			"INSERT INTO links (id, url) VALUES ($1, $2) RETURNING id, create_ts",
			link.Id,
			link.Url,
		).Scan(&link.Id, &link.CreateTimestamp)
	} else {
		err = conn.QueryRow(context.Background(),
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

func GetLink(id string) *types.Link {
	link := new(types.Link)
	err := conn.QueryRow(context.Background(),
		"SELECT id, url, create_ts, counter FROM links WHERE id = $1",
		id,
	).Scan(&link.Id, &link.Url, &link.CreateTimestamp, &link.Counter)
	if err != nil {
		log.Printf("Failed to retrieve link %s: %s\n", id, err)
		return nil
	}

	return link
}

func IncrementLinkCounter(id string) int64 {
	var counter int64
	err := conn.QueryRow(context.Background(),
		"UPDATE links SET counter = counter + 1 WHERE id = $1 RETURNING counter",
		id,
	).Scan(&counter)

	if err != nil {
		log.Printf("Failed to increment link counter: %s\n", err)
		return counter
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

package db

import (
	"context"
	"log"
	"math/rand"
	"time"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"u9k/config"
	"u9k/models"
	"u9k/types"
)

// shared connection pool
var pool *pgxpool.Pool

// The error returned from GetEmailSubscribeLink when a user (email address)
// has unsubscribed from emails.
var ErrEmailUnsubscribed = errors.New("User is unsubscribed from emails")

func Init(forceVersion int) {
	// set up connection configuration
	conf, err := pgxpool.ParseConfig(config.DbConnUrl)
	if err != nil {
		log.Fatalf("Failed to configure the database: %s\n", err)
	}

	// connect to the database

	// use "pgxpool.Connect()" instead of "pgx.Connect()" because the vanilla driver is not safe
	// for concurrent connections (unlike the other Golang SQL drivers)
	// https://github.com/jackc/pgx/wiki/Getting-started-with-pgx
	pool, err = pgxpool.ConnectConfig(context.Background(), conf)
	if err != nil {
		log.Fatalf("Failed to connect to database: %s\n", err)
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
		"INSERT INTO files (filename, filetype, filesize, expire) VALUES ($1, $2, $3, $4) RETURNING id, create_ts",
		file.Name,
		file.Type,
		file.Size,
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
		"SELECT id, filename, filetype, filesize, create_ts, counter, expire, emails_sent FROM files WHERE id = $1",
		id,
	).Scan(&file.Id, &file.Name, &file.Type, &file.Size, &file.CreateTimestamp, &file.Counter, &expire, &file.EmailsSent)
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
		"SELECT id, filename, filetype, create_ts, counter, expire, emails_sent FROM files WHERE create_ts + expire < NOW()",
	)
	if err != nil {
		return files, err
	}
	defer rows.Close()

	for rows.Next() {
		var expire time.Duration
		var file models.File
		err := rows.Scan(&file.Id, &file.Name, &file.Type, &file.CreateTimestamp, &file.Counter, &expire, &file.EmailsSent)
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

func IncreaseFileEmailsSent(id string, n int64) (int64, error) {
	err := pool.QueryRow(context.Background(),
		"UPDATE files SET emails_sent = emails_sent + $1 WHERE id = $2 RETURNING emails_sent",
		n,
		id,
	).Scan(&n)
	if err != nil {
		log.Printf("Failed to increase emails_sent to %d for %s: %s\n", n, id, err)
	}

	return n, err
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

func GetEmail(subscribeLink string) (m types.Email, err error) {
	err = pool.QueryRow(context.Background(),
		"SELECT address, subscribe_link, unsubscribed FROM email_list WHERE subscribe_link = $1",
		subscribeLink,
	).Scan(&m.AddressHash, &m.SubscribeLink, &m.Unsubscribed)

	if err != nil {
		log.Printf("Failed to get email: %s\n", err)
		return m, err
	}

	return m, nil
}

func SaveEmail(m *types.Email) (err error) {
	if m.SubscribeLink == "" {
		// create a new entry in db
		err = pool.QueryRow(context.Background(),
			"INSERT INTO email_list (address, unsubscribed) VALUES ($1, $2) RETURNING subscribe_link, unsubscribed",
			m.AddressHash,
			m.Unsubscribed,
		).Scan(&m.SubscribeLink, &m.Unsubscribed)
	} else {
		// update the existing entry
		err = pool.QueryRow(context.Background(),
			"UPDATE email_list SET unsubscribed = $2 WHERE subscribe_link = $1 RETURNING subscribe_link, unsubscribed",
			m.SubscribeLink,
			m.Unsubscribed,
		).Scan(&m.SubscribeLink, &m.Unsubscribed)
	}
	if err != nil {
		log.Printf("Failed to update email: %s\n", err)
		return err
	}

	return nil
}

// GetEmailSubscribeLink returns the unique subscription ID for the specified email address
// If the email address does not exist in the database yet, it will be created
// If the user unsubscribed from emails, returns ErrEmailUnsubscribed
func GetEmailSubscribeLink(emailTo string) (subscribeLink string, err error) {
	m := types.Email{
		AddressHash: hashSecret(emailTo),
	}
	err = pool.QueryRow(context.Background(),
		"SELECT subscribe_link, unsubscribed FROM email_list WHERE address = $1",
		m.AddressHash,
	).Scan(&m.SubscribeLink, &m.Unsubscribed)
	if err == pgx.ErrNoRows {
		// need to create a new entry in the database
		err = SaveEmail(&m)
	}
	if err != nil {
		log.Printf("Failed to generate subscribe link: %s\n", err)
		return "", err
	}

	if m.Unsubscribed {
		return "", ErrEmailUnsubscribed
	}

	return m.SubscribeLink, nil
}

func hashSecret(secret string) string {
	// bcrypt hashes and salts the secret
    hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }

	return string(hashed)
}

func compareSecret(hashed, plain string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false
	} else {
		log.Printf("Failed to compareSecret: %s", err)
		return false
	}

	return true
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

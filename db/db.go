package db

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"

	"u9k/types"
)

// shared connection handler
var conn *pgx.Conn

// configuration
const letterBytes = "abcdefghijklmnopqrstuvwxyz23456789" // omit 1 and 0 for readability
const linkLength = 6

func InitDBConnection(url string) {
    rand.Seed(time.Now().UnixNano())

	// set up connection configuration
	config, err := pgx.ParseConfig(url)
	if err != nil {
		log.Fatal("error configuring the database: ", err)
	}

	// connect to the database
	conn, err = pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	//defer conn.Close(context.Background())

	// check if connection is working
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to communcate with database: ", url)
	}

	log.Printf("Connected to database %s\n", url)
}


func StoreLink(link *types.Link) (string) {
	log.Printf("Storing Link: %s\n", link.Url)

	if _, err := conn.Exec(context.Background(),
		"INSERT INTO links (id, url) VALUES ($1, $2)", // TODO: RETURNING id
		link.Id,
		link.Url,
	); err != nil {
		log.Printf("Failed to insert link: %s\n", err)
		return ""
	}

	return link.Id // TODO: RETURNING id
}

func GetLink(id string) *types.Link {
	res := conn.QueryRow(context.Background(),
		"SELECT id, url, create_ts FROM links WHERE id = $1",
		id,
	)

	link := new(types.Link)
	err := res.Scan(&link.Id, &link.Url, &link.CreateTimestamp)
	if err != nil {
		log.Printf("Failed to retrieve link %s: %s\n", id, err)
		return nil
	}

	return link
}

func GenerateLinkId() (id string) {
	return randStringBytesRmndr(linkLength)
}

// from https://stackoverflow.com/a/31832326
func randStringBytesRmndr(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
    }
    return string(b)
}

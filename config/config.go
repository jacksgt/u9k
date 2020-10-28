package config

import (
	"log"
	"os"
)

const Version = "v0.10.1"

func Init() {
	baseUrl := os.Getenv("U9K_BASE_URL")
	if baseUrl == "" {
		baseUrl = "http://localhost:3000/"
	}
	// trim trailing slash
	if baseUrl[len(baseUrl)-1] == '/' {
		baseUrl = baseUrl[:len(baseUrl)-1]
	}
	BaseUrl = baseUrl

	listenAddr := os.Getenv("U9K_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("U9K_PORT")
		if port == "" {
			port = "3000"
		}
	}
	HttpListenAddr = listenAddr + ":" + port

	dbConnUrl := os.Getenv("U9K_DB_CONN_URL")
	if dbConnUrl == "" {
		dbConnUrl = "postgresql://root@localhost:26257/u9k?sslmode=disable"
	}
	DbConnUrl = dbConnUrl

	S3Region = os.Getenv("U9K_S3_REGION")
	if S3Region == "" {
		S3Region = "us-east-1"
	}

	S3AccessKey = os.Getenv("U9K_S3_ACCESS_KEY")
	if S3AccessKey == "" {
		log.Fatalf("No U9K_S3_ACCESS_KEY specified. Aborting.\n")
	}

	S3SecretKey = os.Getenv("U9K_S3_SECRET_KEY")
	if S3SecretKey == "" {
		log.Fatalf("No U9K_S3_SECRET_KEY specified. Aborting.\n")
	}

	S3Bucket = os.Getenv("U9K_S3_BUCKET")
	if S3Bucket == "" {
		S3Bucket = "u9k-dev"
	}

	S3Endpoint = os.Getenv("U9K_S3_ENDPOINT")
	// defaults to AWS

	SmtpHostPort = os.Getenv("U9K_SMTP_HOSTPORT")
	SmtpUser = os.Getenv("U9K_SMTP_USER")
	SmtpPassword = os.Getenv("U9K_SMTP_PASSWORD")
	if SmtpHostPort == "" || SmtpUser == "" || SmtpPassword == "" {
		log.Printf("Warning: no SMTP configuration found in environment, email sending disabled")
		SmtpDisabled = true
	}
}

var BaseUrl string
var HttpListenAddr string
var DbConnUrl string
var S3Region string
var S3AccessKey string
var S3SecretKey string
var S3Bucket string
var S3Endpoint string
var SmtpHostPort string
var SmtpUser string
var SmtpPassword string
var SmtpDisabled bool

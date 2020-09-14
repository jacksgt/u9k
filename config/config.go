package config

import (
	"log"
	"os"
)

func Init() {
	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		baseUrl = "http://localhost:3000/"
	}
	if baseUrl[len(baseUrl)-1] != '/' {
		baseUrl += "/"
	}
	BaseUrl = baseUrl

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	HttpListenAddr = listenAddr + ":" + port

	dbConnUrl := os.Getenv("DB_CONN_URL")
	if dbConnUrl == "" {
		dbConnUrl = "postgresql://root@localhost:26257/u9k?sslmode=disable"
	}
	DbConnUrl = dbConnUrl

	S3Region = os.Getenv("S3_REGION")
	if S3Region == "" {
		S3Region = "us-east-1"
	}

	S3AccessKey = os.Getenv("S3_ACCESS_KEY")
	if S3AccessKey == "" {
		log.Fatalf("No S3_ACCESS_KEY specified. Aborting.\n")
	}

	S3SecretKey = os.Getenv("S3_SECRET_KEY")
	if S3SecretKey == "" {
		log.Fatalf("No S3_SECRET_KEY specified. Aborting.\n")
	}

	S3Bucket = os.Getenv("S3_BUCKET")
	if S3Bucket == "" {
		S3Bucket = "u9k-dev"
	}

	S3Endpoint = os.Getenv("S3_ENDPOINT")
}

var BaseUrl string
var HttpListenAddr string
var DbConnUrl string
var S3Region string
var S3AccessKey string
var S3SecretKey string
var S3Bucket string
var S3Endpoint string

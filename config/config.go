package config

import (
	"os"
)

func Init() {
	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		baseUrl = "http://localhost:3000/"
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
}

var BaseUrl string
var HttpListenAddr string
var DbConnUrl string

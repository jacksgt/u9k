package main

import (
    "fmt"
    "u9k/api"
	"u9k/db"
	"u9k/config"
)

func main() {
	fmt.Println("Launching U9K server ...")
	config.Init()
	db.InitDBConnection()
	api.Init()
}

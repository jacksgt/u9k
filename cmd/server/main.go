package main

import (
	"flag"
	"fmt"
	"u9k/api"
	"u9k/config"
	"u9k/db"
)

func main() {
	fmt.Println("Launching U9K server ...")
	forceMigrationVersion := flag.Int("forceMigrationVersion", 0, "Sets a migration version and resets the dirty state")
	flag.Parse()
	config.Init()
	db.InitDBConnection(*forceMigrationVersion)
	api.Init()
}

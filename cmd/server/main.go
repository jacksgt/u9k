package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"u9k/api"
	"u9k/api/render"
	"u9k/config"
	"u9k/db"
	"u9k/schedules"
	"u9k/storage"
)

func main() {
	fmt.Println("Launching U9K server ...")

	// make sure rand package is properly seeded everywhere
	rand.Seed(time.Now().UnixNano())

	// parse CLI flags
	forceMigrationVersion := flag.Int("forceMigrationVersion", 0, "Sets a migration version and resets the dirty state")
	reloadTemplates := flag.Bool("reloadTemplates", false, "Reload HTML templates with each request")
	flag.Parse()

	// initialize components (these function will log.Fatal on error)
	config.Init()
	db.Init(*forceMigrationVersion)
	storage.Init()
	schedules.Init()
	render.Init(*reloadTemplates)
	api.Init()
}

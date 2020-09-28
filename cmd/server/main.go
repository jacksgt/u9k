package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"u9k/api"
	"u9k/api/render"
	"u9k/config"
	"u9k/db"
	"u9k/schedules"
	"u9k/storage"
)

func main() {
	// make sure rand package is properly seeded everywhere
	rand.Seed(time.Now().UnixNano())

	// parse CLI flags
	runHealthCheck := flag.Bool("runHealthCheck", false, "Run a healthcheck against the endpoint specified via U9K_LISTEN_ADDR and U9K_PORT")
	forceMigrationVersion := flag.Int("forceMigrationVersion", 0, "Sets a migration version and resets the dirty state")
	reloadTemplates := flag.Bool("reloadTemplates", false, "Reload HTML templates with each request")
	flag.Parse()

	if *runHealthCheck {
		healthcheck()
		log.Println("Healthy")
		return
	}

	// initialize components (these function will log.Fatal on error)
	log.Println("Launching U9K server ...")
	config.Init()
	db.Init(*forceMigrationVersion)
	storage.Init()
	schedules.Init()
	render.Init(*reloadTemplates)
	api.Init()
}

// performs a GET request against /health/ and returns an error, if any
func healthcheck() {
	url := fmt.Sprintf("http://%s:%s/health/", os.Getenv("U9K_LISTEN_ADDR"), os.Getenv("U9K_PORT"))
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		log.Fatalf("GET %s returned %s", url, resp.Status)
	}
}

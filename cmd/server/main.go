package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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
	genericInit()
	db.Init(*forceMigrationVersion)
	storage.Init()
	err := storage.Check()
	if err != nil {
		log.Fatalf("Failed to initialize storage backend: %s", err)
	}
	schedules.Init()
	render.Init(*reloadTemplates)
	api.Init()
}

// performs generic initialization functions
// exits the main thread on error
func genericInit() {
	// make sure we have a working tempdir, because:
	// os.TempDir(): The directory is neither guaranteed to exist nor have accessible permissions.
	tempDir := os.TempDir()
	if err := os.MkdirAll(tempDir, 1777); err != nil {
		log.Fatalf("Failed to create temporary directory %s: %s", tempDir, err)
	}
	tempFile, err := ioutil.TempFile("", "genericInit_")
	if err != nil {
		log.Fatalf("Failed to create tempFile: %s", err)
	}
	_, err = fmt.Fprintf(tempFile, "Hello, World!")
	if err != nil {
		log.Fatalf("Failed to write to tempFile: %s", err)
	}
	if err := tempFile.Close(); err != nil {
		log.Fatalf("Failed to close tempFile: %s", err)
	}
	if err := os.Remove(tempFile.Name()); err != nil {
		log.Fatalf("Failed to delete tempFile: %s", err)
	}
	log.Printf("Using temporary directory %s", tempDir)
}

// performs a GET request against /health/ and returns an error, if any
func healthcheck() {
	url := fmt.Sprintf("http://%s:%s/health/", os.Getenv("U9K_LISTEN_ADDR"), os.Getenv("U9K_PORT"))
	// set low timeout on the health endpoint to detect responsiveness
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		log.Fatalf("GET %s returned %s", url, resp.Status)
	}
}

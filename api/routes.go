package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"

	"u9k/config"
)

func Init() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	staticFS := http.FileServer(http.Dir("./static/"))
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static", staticFS).ServeHTTP(w, r)
	})

	// to avoid lookups to the database which result in 404 anyway
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		return
	})
	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		// do not index anything else than the main site
		fmt.Fprintf(w, "User-agent: *\nDisallow: /\nAllow: /index.html\n")
		return
	})

	r.Group(func(r chi.Router) {
		// limit endpoints in this group to one request per second
		r.Use(httprate.Limit(1, 1*time.Second))
		r.Post("/link/", postLinkHandler)
		r.Post("/file/", postFileHandler)
	})
	r.Get("/link/{linkId}", previewLinkHandler)
	r.Get("/file/{fileId}", getFileHandler)
	r.Get("/{linkId}", getLinkHandler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	r.Get("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	log.Printf("HTTP Server listening on %s\n", config.HttpListenAddr)
	http.ListenAndServe(config.HttpListenAddr, r)
}

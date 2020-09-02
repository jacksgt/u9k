package main

import (
    "fmt"
    "u9k/api"
	"u9k/db"
	"net/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func initHttpServer() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	staticFS := http.FileServer(http.Dir("./static/"))
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static", staticFS).ServeHTTP(w, r)
	})
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// to avoid lookups to the database which result in 404 anyway
		return
	})
	r.Post("/link/", api.PostLinkHandler)
	r.Get("/{linkId}", api.GetLinkHandler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	http.ListenAndServe(":3000", r)
}

func main() {
	fmt.Println("")
	db.InitDBConnection("postgresql://root@localhost:26257/u9k?sslmode=disable")
	initHttpServer()
}

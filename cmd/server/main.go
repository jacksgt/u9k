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
	r.Get("/welcome", api.GetWelcome)
	r.Post("/link/", api.PostLinkHandler)
	r.Get("/{linkId}", api.GetLinkHandler)
	http.ListenAndServe(":3000", r)
}

func main() {
	fmt.Println("")
	db.InitDBConnection("postgresql://root@localhost:26257/u9k?sslmode=disable")
	initHttpServer()
}

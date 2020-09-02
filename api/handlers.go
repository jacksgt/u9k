package api

import (
	"log"
	"fmt"
	"net/http"
	"u9k/types"
	"u9k/db"
	"u9k/api/render"

	"github.com/go-chi/chi"
)

const baseUrl = "http://localhost:3000/"


func GetWelcome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome"))
	log.Printf("")
}

func PostLinkHandler(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")
	if url == "" {
		httpError(w, "Missing field 'url' in request", 400)
		return
	}

	if ! isValidUrl(url) {
		httpError(w, "Data in 'url' field is not a valid URL", 400)
		return
	}

	link := new(types.Link)
	link.Url = url
	link.Id = db.GenerateLinkId()

	id := db.StoreLink(link)

	fmt.Fprintf(w, "%s%s\n", baseUrl, id)
	return
}

func GetLinkHandler(w http.ResponseWriter, r *http.Request) {
	linkId := chi.URLParam(r, "linkId")
	if linkId == "" {
		httpError(w, "Not Found", 404)
		return
	}

	link := db.GetLink(linkId)
	if link == nil {
		httpError(w, "Not Found", 404)
		return
	}

	render.RedirectLink(w, r, link.Url)
	return
}

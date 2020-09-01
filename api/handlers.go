package api

import (
	"fmt"
	"net/http"
	"u9k/types"
	"u9k/db"

	"github.com/go-chi/chi"
)

const baseUrl = "http://localhost:3000/"


func GetWelcome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome"))
}

func PostLinkHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		httpError(w, "Missing field 'data' in request", 400)
		return
	}

	link := new(types.Link)
	link.Url = r.FormValue("url")
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

	fmt.Fprintf(w, "%s\n", link.Url)
	return
}

package api

import (
	"fmt"
	"net/http"

	"u9k/api/render"
	"u9k/db"
	"u9k/models"

	"github.com/go-chi/chi"
)

func postLinkHandler(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")
	if url == "" {
		httpError(w, "Missing field 'url' in request", 400)
		return
	}

	if !isValidUrl(url) {
		httpError(w, "Data in 'url' field is not a valid URL", 400)
		return
	}

	link := new(models.Link)
	link.Url = url
	shortLink := r.PostFormValue("link")
	if shortLink != "" {
		if !isValidLinkId(shortLink) {
			httpError(w, "Some characters in 'link' field are not allowed", 400)
			return
		}
		link.Id = shortLink
	}

	err := db.StoreLink(link)
	if err != nil {
		httpError(w, "Internal Server Error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "%s\n", link.ToJson())
	return
}

func getLinkHandler(w http.ResponseWriter, r *http.Request) {
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

	db.IncrementLinkCounter(linkId)

	render.RedirectTo(w, r, link.Url)
	return
}

func previewLinkHandler(w http.ResponseWriter, r *http.Request) {
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

	render.RedirectLinkPage(w, r, link)
	return
}

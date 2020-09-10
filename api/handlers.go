package api

import (
	"fmt"
	"net/http"

	"u9k/api/render"
	"u9k/config"
	"u9k/db"
	"u9k/types"

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

	link := new(types.Link)
	link.Url = url
	shortLink := r.PostFormValue("link")
	if shortLink != "" {
		if !isValidLinkId(shortLink) {
			httpError(w, "Some characters in 'link' field are not allowed", 400)
			return
		}
		link.Id = shortLink
	}

	id := db.StoreLink(link)
	if id == "" {
		httpError(w, "Internal Server Error", 500)
		return
	}

	fmt.Fprintf(w, "%s%s\n", config.BaseUrl, id)
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

	render.RedirectLink(w, r, link.Url)
	return
}

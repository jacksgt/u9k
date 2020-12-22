package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"

	"u9k/api/render"
	"u9k/config"
)

func Init() {
	r := chi.NewRouter()

	// static files
	staticFS := http.FileServer(http.Dir("./static/"))
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static", staticFS).ServeHTTP(w, r)
	})
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/static/icons/favicon.ico")
		return
	})
	// to avoid lookups to the database which result in 404 anyway
	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		// do not index anything else than the main site
		fmt.Fprintf(w, "User-agent: *\nDisallow: /\nAllow: /index.html\nAllow: /$\n")
		return
	})

	r.Get("/health/", healthHandler)
	r.Group(func(r chi.Router) {
		// endpoints in this grouped are logged
		r.Use(middleware.Logger)

		r.Group(func(r chi.Router) {
			// limit endpoints in this group to one request per second
			r.Use(httprate.Limit(1, 1*time.Second))
			r.Post("/link/", postLinkHandler)
			r.Post("/file/", postFileHandler)
			r.Post("/file/{fileId}/email", sendFileEmailHandler)
		})
		r.Get("/ip/", getIpHandler)
		r.Get("/email/{emailLink}", getEmailHandler)
		r.Post("/email/{emailLink}", postEmailHandler)
		r.Get("/link/{linkId}", previewLinkHandler)
		r.Get("/file/{fileId}/raw/{filename}", rawFileHandler)
		r.Get("/file/{fileId}", getFileHandler)
		r.Get("/{linkId}", getLinkHandler)

	})
	r.Get("/video-audio-test/", videoAudioTestHandler)
	r.Get("/", indexHandler)
	r.Get("/index.html", indexHandler)

	log.Printf("HTTP Server listening on %s\n", config.HttpListenAddr)
	http.ListenAndServe(config.HttpListenAddr, r)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	render.Index(w)
}

func videoAudioTestHandler(w http.ResponseWriter, r *http.Request) {
	render.VideoAudio(w)
}

func subscribeUrl(subscribeLink string) string {
	return fmt.Sprintf("%s/email/%s", config.BaseUrl, subscribeLink)
}

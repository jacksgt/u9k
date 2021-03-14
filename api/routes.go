package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/cors"

	"u9k/api/render"
	"u9k/config"
)

func Init() {
	r := chi.NewRouter()
	r.Use(securityHeaders)

	// static files
	staticFS := http.FileServer(http.Dir("./static/"))
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		// allow static resources to be embedded on our site
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
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

		// enable CORS
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{config.BaseUrl},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))

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

// HTTP middleware for setting secure headers
func securityHeaders(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  // from https://geekflare.com/http-header-implementation/
	  // block XSS attempts:
	  w.Header().Set("X-XSS-Protection", "1; mode=block")
	  // do not allow embedding as frame / iframe / embed / object:
	  w.Header().Set("X-Frame-Options", "DENY")
	  // do not automatically try to detect MIME types:
	  w.Header().Set("X-Content-Type-Options", "nosniff")
	  // all content comes from the site's own origin + inline stuff:
	  // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
	  // https://scotthelme.co.uk/content-security-policy-an-introduction/
	  w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; object-src 'self'; base-uri 'self'; form-action 'self'; media-src 'self' blob:;")
	  // do not send referrer headers when navigating to other sites:
	  w.Header().Set("Referrer-Policy", "same-origin")

	  // pass the request onto the next handler in the chain
	  next.ServeHTTP(w, r)
  })
}

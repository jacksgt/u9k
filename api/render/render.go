package render

import (
	"fmt"
	"net/http"
	"time"

	"u9k/models"
)

func RedirectLinkPage(w http.ResponseWriter, r *http.Request, link *models.Link) {
	fmt.Fprintf(w, "The link %s points to %s<br>\n", link.ExportLink(), link.Url)
	fmt.Fprintf(w, "Created on %s, used %d times since then<br>\n", link.CreateTimestamp.Format("2006-01-02"), link.Counter)
}

func RedirectLink(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(302)

	// for text clients
	fmt.Fprintf(w, "%s\n", url)
}

func PreviewFile(w http.ResponseWriter, r *http.Request, f *models.File) {
	fmt.Fprintf(w, "Filename: %s\n Link: %s?raw=true\n Downloads: %d\n Expires at: %s\n", f.Name, f.ExportLink(), f.Counter, f.CreateTimestamp.Add(time.Duration(f.Expire)))
}

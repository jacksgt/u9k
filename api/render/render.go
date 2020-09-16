package render

import (
	"fmt"
	"net/http"

	"u9k/config"
	"u9k/types"
)

func RedirectLinkPage(w http.ResponseWriter, r *http.Request, link *types.Link) {
	fmt.Fprintf(w, "The link %s%s points to %s<br>\n", config.BaseUrl, link.Id, link.Url)
	fmt.Fprintf(w, "Created on %s, used %d times since then<br>\n", link.CreateTimestamp.Format("2006-01-02"), link.Counter)
}

func RedirectLink(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(302)

	// for text clients
	fmt.Fprintf(w, "%s\n", url)
}

func PreviewFile(w http.ResponseWriter, r *http.Request, f *types.File) {
	f.ExportToJson() // generate link
	fmt.Fprintf(w, "Filename: %s<br>\n Link: %s?raw=true<br>\n Downloads: %d\n", f.Name, f.Link, f.Counter)
}

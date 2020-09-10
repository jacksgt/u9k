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
	return
}

func RedirectLink(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(302)

	// for text clients
	fmt.Fprintf(w, "%s\n", url)
	return
}

package render

import (
	"fmt"
	"net/http"
)

func RedirectLinkPage(w http.ResponseWriter, r *http.Request, url string) {

}

func RedirectLink(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(302)

	// for text clients
	fmt.Fprintf(w, "%s\n", url)
	return
}

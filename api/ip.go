package api

import (
	"log"
	"net/http"
)

func getIpHandler(w http.ResponseWriter, r *http.Request) {
	// does not nothing except collecting info for now

	log.Printf("%#v\n%#v", r, r.URL)

	return
}

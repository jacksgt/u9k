package api

import (
	"fmt"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK - Healthy")
	// TODO:
	// check database
	// check storage
	// check templates
}

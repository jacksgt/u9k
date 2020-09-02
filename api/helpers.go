package api

import (
	"fmt"

	"net/http"
	"net/url"
)

// same interface as http.Error()
func httpError(w http.ResponseWriter, message string, code int){
	// Display the the footer ("Contact admin etc.") only when theres a server error
	// footer := "none"
	// if code >= 500 {
	// 	footer = "block"
	// }

	// writer.WriteHeader(code)
	// Templates["error.html"].Execute(writer,
	// 	map[string]interface{}{
	// 		"APPNAME": Config.App.Name,
	// 		"ERRORCODE": strconv.Itoa(code),
	// 		"ERRORMESSAGE": message,
	// 		"FOOTER": footer,
	// 	})

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "%d - %s\n", code, message)
	return
}

// adapted from https://stackoverflow.com/a/55551215
func isValidUrl(str string) (bool) {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
)

const validLinkRegex = "[a-zA-Z0-9-_]{6,}"

// same interface as http.Error()
func httpError(w http.ResponseWriter, message string, code int) {
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

func httpError2(w http.ResponseWriter, r *http.Request, code int, messages ...string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)

	message := httpCodes[code] + "\n"
	for _, m := range messages {
		message += m + "\n"
	}

	fmt.Fprintf(w, "%d - %s\n", code, message)
}

// adapted from https://stackoverflow.com/a/55551215
func isValidUrl(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isValidLinkId(str string) bool {
	match, err := regexp.MatchString(validLinkRegex, str)
	if err != nil {
		log.Printf("Regex %s error: %s\n", err)
	}
	// in case of error, "match" is always "false"
	return match
}

func getFormFile(name string, w http.ResponseWriter, r *http.Request) ([]byte, string) {
	var data []byte
	var uploadFileName string

	// parse uploaded data
	err := r.ParseMultipartForm(10000000) // 10 MB in memory, rest on disk
	if err != nil {
		log.Printf("Failed to parse form: %s\n", err)
		return data, uploadFileName
	}

	// get file from form
	fileHeaders := r.MultipartForm.File[name]
	if fileHeaders == nil || len(fileHeaders) < 1 {
		log.Printf("No files found in request.\n")
		return data, uploadFileName
	}

	// open and read the corresponding file into memory
	fd, err := fileHeaders[0].Open()
	if err != nil {
		log.Printf("Failed to open file from client: %s\n", err)
		return data, uploadFileName
	}
	uploadFileName = fileHeaders[0].Filename
	data, err = ioutil.ReadAll(fd)
	if err != nil {
		log.Printf("helper.go: Failed to read file from client: %s\n", err)
		return data, uploadFileName
	}

	return data, uploadFileName
}

func extractFormFileHeader(name string, r *http.Request) *multipart.FileHeader {
	// parse uploaded data
	err := r.ParseMultipartForm(10000000) // 10 MB in memory, rest on disk
	if err != nil {
		log.Printf("Failed to parse form: %s\n", err)
		return nil
	}

	// get file from form
	fileHeaders := r.MultipartForm.File[name]
	if fileHeaders == nil || len(fileHeaders) < 1 {
		log.Printf("No files found in request.\n")
		return nil
	}

	return fileHeaders[0]
}

// extrated from https://github.com/emirozer/go-helpers/
func stringInSlice(str string, slice []string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// adapted from https://golangcode.com/get-the-content-type-of-file/
func getFileContentType(r io.Reader) string {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := r.Read(buffer)
	if err != nil {
		log.Printf("Failed to detect Content-Type: %s\n", err)
		return "application/octet-stream"
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)
	return contentType
}

var httpCodes = map[int]string{
	200: "OK",
	400: "Bad Request",
	404: "Not Found",
	500: "Internal Server Error",
}

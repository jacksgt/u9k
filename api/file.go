package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"u9k/api/render"
	"u9k/db"
	"u9k/models"
	"u9k/storage"
	"u9k/types"

	"github.com/go-chi/chi"
)

const MAX_EMAILS = 1

func postFileHandler(w http.ResponseWriter, r *http.Request) {
	// get the file from the form (name should be "file")
	fh := extractFormFileHeader("file", r)
	if fh == nil {
		httpError2(w, r, 400, "No file found in POST request")
		return
	}

	// open filehandle
	fd, err := fh.Open()
	if err != nil {
		httpError2(w, r, 400, "Failed to read uploaded file")
		return
	}
	defer fd.Close()

	expireStr := r.PostFormValue("expire")
	if expireStr == "" {
		expireStr = "168h" // 1 week
	}

	expire, err := time.ParseDuration(expireStr)
	if err != nil {
		httpError2(w, r, 400, "Invalid format in 'expire' field")
		return
	}
	if expire > time.Duration(time.Hour*24*366) { // maximum one year
		httpError2(w, r, 400, "Expiry time too large (max. one year)")
		return
	}

	file := models.File{
		Name:   fh.Filename,
		Size:   fh.Size,
		Type:   getFileContentType(fd),
		Expire: types.Duration(expire),
	}

	// save metadata in the DB
	err = db.StoreFile(&file)
	if err != nil {
		httpError(w, "Internal Server Error", 500)
		return
	}

	// store data in storage backend
	key := storage.FileKey(file.Id, file.Name)
	err = storage.StoreFileStream(fd, key, file.Type)
	if err != nil {
		httpError2(w, r, 500)
		return
	}

	// TODO: need to delete temporary multipart form file after uploading

	// TODO: should return 201 - Created, set Location header
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "%s\n", file.ToJson())
	return
}

func getFileHandler(w http.ResponseWriter, r *http.Request) {
	file := db.GetFile(chi.URLParam(r, "fileId"))
	if file == nil {
		httpError(w, "Not Found", 404)
		return
	}

	err := r.ParseForm()
	if err != nil {
		httpError(w, "Bad Request", 400)
		return
	}

	// supported for backwards compatibility
	if r.FormValue("raw") == "true" {
		render.RedirectTo(w, r, file.RawLink())
		return
	}

	render.PreviewFile(w, r, file)
	return
}

func rawFileHandler(w http.ResponseWriter, r *http.Request) {
	file := db.GetFile(chi.URLParam(r, "fileId"))
	if file == nil {
		httpError(w, "Not Found", 404)
		return
	}

	// download file from backend
	key := storage.FileKey(file.Id, file.Name)
	data, err := storage.GetFile(key)
	if err != nil {
		httpError(w, "Internal Server Error", 500)
		return
	}

	// run in the background
	go db.IncrementCounter("file", file.Id)

	// set content-type header manually, so ServeContent() won't try to guess it
	w.Header().Set("Content-Type", file.Type)

	// serve to client
	rs := bytes.NewReader(data)
	http.ServeContent(w, r, file.Name, file.CreateTimestamp, rs)
}

func sendFileEmailHandler(w http.ResponseWriter, r *http.Request) {
	toEmail := r.PostFormValue("to_email")
	fromName := r.PostFormValue("from_name")
	if toEmail == "" || fromName == "" {
		httpError(w, "Bad Request - missing to_email or from_name field", 400)
		return
	}

	fileId := chi.URLParam(r, "fileId")
	file := db.GetFile(fileId)
	if file == nil {
		httpError(w, "Not Found", 404)
		return
	}

	// check if we have already sent emails for this file
	if file.EmailsSent >= MAX_EMAILS {
		httpError(w, "Too Many Requests", 429)
		return
	}

	// check if the recipient has unsubscribed from emails
	subscribeLink, err := db.GetEmailSubscribeLink(toEmail)
	if err == db.ErrEmailUnsubscribed {
		log.Printf("Aborting email request, recipient unsubscribed from emails")
		httpError(w, "Bad Request", 400)
		return
	}
	if err != nil {
		log.Printf("Not allowed to send emails to %s\n", toEmail)
		// TODO: implement a better response and make it visible to the user
		httpError(w, "Internal Server Error", 500)
		return
	}

	ew, err := render.FileMail(fromName, file, subscribeUrl(subscribeLink))
	if err != nil {
		log.Printf("%s", err)
		return
	}

	// for debugging the HTML template:
	// fmt.Fprintf(w, "%s", ew.HtmlBody)
	// fmt.Fprintf(w, "%v", ew)

	// TODO: return 202 - Accepted, run the rest in the background
	err = ew.SendTo(toEmail)
	if err != nil {
		log.Printf("Failed to send email to %s: %s", toEmail, err)
		httpError(w, "Internal Server Error", 500)
		return
	}

	db.IncreaseFileEmailsSent(file.Id, 1)

	fmt.Fprintf(w, "OK\n")
}

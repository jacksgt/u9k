package api

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"u9k/api/render"
	"u9k/db"
	"u9k/models"
	"u9k/storage"
	"u9k/types"

	"github.com/go-chi/chi"
)

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

	file := new(models.File)
	file.Name = fh.Filename
	file.Size = fh.Size
	file.Type = getFileContentType(fd)
	file.Expire = types.Duration(expire)

	// save metadata in the DB
	err = db.StoreFile(file)
	if err != nil {
		httpError(w, "Internal Server Error", 500)
		return
	}

	// store data in storage backend
	err = storage.StoreFileStream(fd, file.Id, file.Type)
	if err != nil {
		httpError2(w, r, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "%s\n", file.ToJson())
	return
}

func getFileHandler(w http.ResponseWriter, r *http.Request) {
	fileId := chi.URLParam(r, "fileId")
	file := db.GetFile(fileId)
	if file == nil {
		httpError(w, "Not Found", 404)
		return
	}

	err := r.ParseForm()
	if err != nil {
		httpError(w, "Bad Request", 400)
		return
	}

	if r.FormValue("raw") == "true" {
		db.IncrementCounter("file", fileId)

		// TODO: maybe it's better to generate a signed URL and redirect to it

		// download file from backend
		data, err := storage.GetFile(file.Id)
		if err != nil {
			httpError(w, "Internal Server Error", 500)
			return
		}

		// serve to client
		rs := bytes.NewReader(data)
		http.ServeContent(w, r, file.Name, file.CreateTimestamp, rs)
		return
	}

	render.PreviewFile(w, r, file)
	return
}

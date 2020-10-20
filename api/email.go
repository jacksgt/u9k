package api

import (
	"net/http"

	"github.com/go-chi/chi"

	"u9k/api/render"
	"u9k/db"
)

func getEmailHandler(w http.ResponseWriter, r *http.Request) {
	subscribeLink := chi.URLParam(r, "emailLink")
	email, err := db.GetEmail(subscribeLink)
	if err != nil {
		httpError(w, "Not Found", 404)
		return
	}

	render.EmailSubscribePage(w, email)
	return
}

func postEmailHandler(w http.ResponseWriter, r *http.Request) {
	subscribeLink := chi.URLParam(r, "emailLink")
	email, err := db.GetEmail(subscribeLink)
	if err != nil {
		httpError(w, "Not Found", 404)
		return
	}

	// get user choice from form input
	email.Unsubscribed = r.PostFormValue("unsubscribe") == "true"

	// save to database
	err = db.SaveEmail(&email)
	if err != nil {
		httpError(w, "Internal Server Error", 500)
		return
	}

	// show the new status to the user
	getEmailHandler(w, r)
}

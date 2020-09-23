package render

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"u9k/models"
)

var reload bool = false

func RedirectLinkPage(w http.ResponseWriter, r *http.Request, link *models.Link) {
	Template(w, "link.html", link)
}

func RedirectLink(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(302)

	// for text clients
	fmt.Fprintf(w, "%s\n", url)
}

func PreviewFile(w http.ResponseWriter, r *http.Request, f *models.File) {
	fmt.Fprintf(w, "Filename: %s\n Link: %s?raw=true\n Downloads: %d\n Expires at: %s\n", f.Name, f.ExportLink(), f.Counter, f.CreateTimestamp.Add(time.Duration(f.Expire)))
}

var templates *template.Template

func Init(reloadTemplates bool) {
	if reloadTemplates {
		reload = true
		return
	}

	loadTemplates()
}

func loadTemplates() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		panic("Failed to load templates: " + err.Error())
	}
	log.Printf("Loaded templates: %s\n", templates.DefinedTemplates())
}

func Template(w http.ResponseWriter, name string, data interface{}) {
	if reload {
		// in development mode, reload templates with each request
		loadTemplates()
	}

	t := templates.Lookup(name)
	if t == nil {
		log.Printf("Failed to find template '%s'\n", name)
		// TODO: httpError
	}

	err := t.Execute(w, data)
	if err != nil {
		log.Printf("Template execution failed: %s\n", err)
		// TODO: httpError
	}
}

func Index(w http.ResponseWriter) {
	Template(w, "index.html", nil)
}

package render

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"u9k/config"
	"u9k/models"
	"u9k/types"
)

var reload bool = false

type M map[string]interface{}

var appConfig M

var templates *template.Template

var globalTemplateFunctions = template.FuncMap{
	"absUrl": func(path string) string {
		/* make sure there is a leading slash */
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return config.BaseUrl + path
	},
	"htmlSafe": func(html string) template.HTML {
		return template.HTML(html)
	},
}

func Init(reloadTemplates bool) {
	// craft another config object so we don't accidentally expose any sensitive data
	appConfig = M{
		"Version": config.Version,
	}

	if reloadTemplates {
		// loads templates before each execution
		reload = true
		return
	}

	loadTemplates()
}

func loadTemplates() {
	var err error
	templates, err = template.New("").Funcs(globalTemplateFunctions).ParseGlob("templates/*.html")
	if err != nil {
		panic("Failed to load templates: " + err.Error())
	}

	// strip the static prefix
	buf := strings.TrimPrefix(templates.DefinedTemplates(), "; defined templates are: ")
	log.Printf("Loaded templates: %s", buf)
}

func execTemplate(w io.Writer, name string, data interface{}) error {
	if reload {
		// in development mode, reload templates with each request
		loadTemplates()
	}

	t := templates.Lookup(name)
	if t == nil {
		err := errors.New(fmt.Sprintf("Failed to find template '%s'", name))
		log.Printf("%s", err)
		return err
	}

	err := t.Execute(w, data)
	if err != nil {
		log.Printf("Template execution failed: %s\n", err)
		return err
	}

	return nil
}

func RedirectLinkPage(w http.ResponseWriter, r *http.Request, link *models.Link) {
	data := M{
		"Link":   link,
		"Config": appConfig,
	}
	execTemplate(w, "link.html", data)
}

func RedirectLink(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(302)

	// write body for text clients
	fmt.Fprintf(w, "%s\n", url)
}

func PreviewFile(w http.ResponseWriter, r *http.Request, f *models.File) {
	data := M{
		"File":   f,
		"Config": appConfig,
	}
	execTemplate(w, "file.html", data)
}

func Index(w http.ResponseWriter) {
	data := M{
		"Config": appConfig,
	}
	execTemplate(w, "index.html", data)
}

func Mail(w http.ResponseWriter, m *types.MailContent) (string, error) {
	data := M{
		"Config": appConfig,
		"Mail":   m,
	}
	execTemplate(w, "mail.html", data)
	// TODO
	return "", nil
}

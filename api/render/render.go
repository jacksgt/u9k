package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"u9k/config"
	"u9k/email"
	"u9k/misc"
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

func RedirectTo(w http.ResponseWriter, r *http.Request, url string) {
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

func VideoAudio(w http.ResponseWriter) {
	data := M{
		"Config": appConfig,
	}
	execTemplate(w, "video-audio-test.html", data)
}

func renderMail(m *types.MailContent) (string, error) {
	data := M{
		"Config": appConfig,
		"Mail":   m,
	}
	var renderedHtml bytes.Buffer
	err := execTemplate(io.Writer(&renderedHtml), "mail.html", data)
	if err != nil {
		return "", nil
	}
	str := renderedHtml.String()
	if str == "" || str == "<nil>" {
		return "", errors.New("Failed to generate email template, received: " + str)
	}
	return str, nil
}

func FileMail(fromName string, f *models.File, subscribeUrl string, message string) (*email.Wrapper, error) {
	var err error
	var ew email.Wrapper
	ew.Subject = fmt.Sprintf("File %s available for download", f.Name)

	var plainMessageText, htmlMessageText string
	message = strings.TrimSpace(message)
	if message != "" {
		plainMessageText = fmt.Sprintf("\nTheir message for you:\n> %s\n", message)
		htmlMessageText = fmt.Sprintf(`<br>Their message for you:<br><b><i>%s</i></b><br>`, template.HTMLEscapeString(message))
	}
	ew.PlainBody = fmt.Sprintf(`
Hello, %s wants to share a file with you!

%s has uploaded "%s" and shared it with you.
The file will be removed in %s.
%s

Click the following link to download the file:
%s

--------------
%s

To unsubscribe from future emails, please visit this link: %s
`,
		fromName, fromName, f.Name, f.PrettyExpiresIn(), plainMessageText, f.ExportLink(), config.BaseUrl, subscribeUrl)

	ew.HtmlBody, err = renderMail(&types.MailContent{
		Summary:      fmt.Sprintf("%s wants to share a file with you", fromName),
		Heading:      fmt.Sprintf("%s wants to share a file with you", fromName),
		ContentHtml:  template.HTML(fmt.Sprintf("%s has uploaded \"%s\" and shared it with you.<br>The file will be removed in %s.<br>%s<br>Click the following link to download the file:<br>", template.HTMLEscapeString(fromName), template.HTMLEscapeString(f.Name), f.PrettyExpiresIn(), htmlMessageText)),
		ButtonUrl:    f.ExportLink(),
		ButtonName:   "Download",
		SubscribeUrl: subscribeUrl,
	})
	if err != nil {
		log.Printf("Failed to render email template: %s", err)
		return nil, err
	}

	return &ew, nil
}

func EmailSubscribePage(w http.ResponseWriter, m types.Email) {
	data := M{
		"Config": appConfig,
		"Email":  m,
	}
	execTemplate(w, "subscribe.html", data)
}

func IPPage(r *http.Request, w http.ResponseWriter, c types.ClientInfo) {
	if acceptsType(r, "text/plain") {
		fmt.Fprintf(w,
			"IP Address: %s\nBrowser: %s\nOperating System:%s\n",
			c.IPAddress,
			c.Browser,
			c.OS,
		)
	} else if acceptsType(r, "application/json") {
		// marshal JSON
		rawJson, err := json.Marshal(c)
		if err != nil {
			log.Printf("IPPage: failed to marshal JSON: %s", err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		fmt.Fprintf(w, "%s", rawJson)
	} else {
		// render html template
		data := M{
			"Config":     appConfig,
			"ClientInfo": c,
		}
		execTemplate(w, "ip.html", data)
	}
}

// acceptsType checks if an HTTP clients accepts a certain mimetype
// For Accept header is unspecified or */*, only allows text/html mimetypes.
// Note: does not support wildcards (yet)
func acceptsType(r *http.Request, mimetype string) bool {
	acceptFields := strings.Split(r.Header.Get("Accept"), ",")

	// discard any weight parameters
	for k := range acceptFields {
		acceptFields[k] = strings.Split(acceptFields[k], ";")[0]
	}

	// if Accept header unspecified, only allow html
	if len(acceptFields) == 0 || misc.StringInSlice("*/*", acceptFields) {
		return mimetype == "text/html"
	}

	// simplistic handling, does not support wildcards
	return misc.StringInSlice(mimetype, acceptFields)
}

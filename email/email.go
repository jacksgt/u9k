package email

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"sync"

	"u9k/config"

	"github.com/domodwyer/mailyak/v3"
)

const (
	mailFromEmail = "hello@u9k.de"
	mailFromName  = "U9K.de"
)

var mailMutex *sync.Mutex = &sync.Mutex{}
var mail *mailyak.MailYak

type Wrapper struct {
	Subject   string
	PlainBody string
	HtmlBody  string
}

func (e Wrapper) SendTo(toEmail string) error {
	if config.SmtpDisabled {
		log.Printf("Email sending disabled, skipping.")
		return nil
	}

	mailMutex.Lock()
	defer mailMutex.Unlock()

	// Create a new email - specify the SMTP host and auth
	if mail == nil {
		smtpHost := strings.Split(config.SmtpHostPort, ":")[0]
		mail = mailyak.New(config.SmtpHostPort, smtp.PlainAuth("", config.SmtpUser, config.SmtpPassword, smtpHost))
	}

	// meta fields
	mail.ClearAttachments()
	mail.To(toEmail)
	mail.From(mailFromEmail)
	mail.FromName(mailFromName)
	mail.Subject(e.Subject)

	// body
	if e.HtmlBody == "" && e.PlainBody == "" {
		// make sure we don't accidentally spam anyone
		return errors.New(fmt.Sprintf("No body specified for email %s to %s", e.Subject, toEmail))
	}
	mail.HTML().Set(e.HtmlBody)
	mail.Plain().Set(e.PlainBody)

	// send
	if err := mail.Send(); err != nil {
		log.Printf("Failed to send email: %s", err)
		return err
	}

	return nil
}

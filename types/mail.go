package types

import "html/template"

type MailContent struct {
	Summary      string // optional
	Heading      string
	ContentHtml  template.HTML
	ButtonUrl    string // optional
	ButtonName   string // optional
	Plaintext    string
	SubscribeUrl string
}

type Email struct {
	Address       string
	SubscribeLink string
	Unsubscribed  bool
}

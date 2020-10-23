package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hako/durafmt"

	"u9k/config"
	"u9k/types"
)

type BaseType interface {
	ExportLink() string
	ToJson() string
}

// BASE DEFINITIONS
type Base struct {
	Id              string    `json:"-"` // omit from JSON
	Link            string    `json:"link"`
	CreateTimestamp time.Time `json:"createTs"`
	Counter         int64     `json:"counter"`
}

// LINK DEFINITIONS
type Link struct {
	Base        // inherit from Base
	Url  string `json:"url"`
}

func (l *Link) ExportLink() string {
	l.Link = config.BaseUrl + "/" + l.Id
	return l.Link
}

func (l *Link) ToJson() string {
	l.ExportLink()
	buf, err := json.Marshal(l)
	if err != nil {
		panic("Error generating JSON:" + err.Error())
	}
	return string(buf)
}

// FILE DEFINITIONS
type File struct {
	Base                      // inherit from Base
	Name       string         `json:"filename"`
	Type       string         `json:"filetype"`
	Size       int64          `json:"filesize"`
	Expire     types.Duration `json:"expire"`
	EmailsSent int16          `json:"emails_sent"`
}

func (f *File) ExportLink() string {
	f.Link = config.BaseUrl + "/file/" + f.Id
	return f.Link
}

func (f *File) RawLink() string {
	return f.ExportLink() + "/raw/" + f.Name
}

func (f *File) ToJson() string {
	f.ExportLink()
	buf, err := json.Marshal(f)
	if err != nil {
		panic("Error generating JSON: " + err.Error())
	}
	return string(buf)
}

func (f *File) PrettyExpiresIn() string {
	d := time.Duration(f.Expire)
	if d.Seconds() == 0 {
		return "Never"
	}

	d = time.Until(f.CreateTimestamp.Add(d))
	// only show the most significant element, e.g. "2 weeks"
	str := durafmt.Parse(d).LimitFirstN(1).String()
	return str
}

func (f *File) PrettyFileSize() string {
	return byteCountIEC(f.Size)
}

// from https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

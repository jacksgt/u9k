package types

import (
	"encoding/json"
	"log"
	"time"

	"u9k/config"
)

// BASE DEFINITIONS
type BaseType interface {
	// Id string
	// Link string
	// CreateTimestamp time.Time
	// Counter int64

	ExportLink() string
	ToJson() string
}

type Base struct {
	Id              string    `json:"-"` // omit from JSON
	Link            string    `json:"link"`
	CreateTimestamp time.Time `json:"create_ts"`
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
		log.Printf("Error generating JSON: %s\n", err)
		return ""
	}
	return string(buf)
}

// FILE DEFINITIONS
type File struct {
	Base        // inherit from Base
	Name string `json:"filename"`
	Type string `json:"filetype"`
	Size int64  `json:"filesize"`
}

func (f *File) ExportLink() string {
	f.Link = config.BaseUrl + "/file/" + f.Id
	return f.Link
}

func (f *File) ToJson() string {
	f.ExportLink()
	buf, err := json.Marshal(f)
	if err != nil {
		log.Printf("Error generating JSON: %s\n", err)
		return ""
	}
	return string(buf)
}

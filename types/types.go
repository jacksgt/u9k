package types

import (
	"encoding/json"
	"fmt"
	"time"

	"u9k/config"
)

type Link struct {
	Id              string    `json:"-"` // omit from JSON output
	Link            string    `json:"link"`
	Url             string    `json:"url"`
	CreateTimestamp time.Time `json:"createTs"`
	Counter         int64     `json:"counter"`
}

func (l *Link) ExportToJson() (string, error) {
	l.Link = fmt.Sprintf("%s%s", config.BaseUrl, l.Id)
	return toJson(l)
}

type File struct {
	Id              string    `json:"-"` // omit from JSON output
	Link            string    `json:"link"`
	Name            string    `json:"filename"`
	Type            string    `json:"filetype"`
	Size            int64     `json:"filesize"`
	CreateTimestamp time.Time `json:"createTs"`
	Counter         int64     `json:"counter"`
}

func (f *File) ExportToJson() (string, error) {
	f.Link = fmt.Sprintf("%sfile/%s", config.BaseUrl, f.Id)
	return toJson(f)
}

func toJson(s interface{}) (string, error) {
	buf, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

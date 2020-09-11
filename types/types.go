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
	buf, err := json.Marshal(l)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

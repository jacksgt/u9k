package api

import (
	"net/http"
	"strings"
	"u9k/api/render"
	"u9k/types"

	"github.com/xojoc/useragent"
)

func getIpHandler(w http.ResponseWriter, r *http.Request) {
	var ip string
	if r.Header.Get("X-Forwarded-For") != "" {
		ip = r.Header.Get("X-Forwarded-For")
	} else if r.Header.Get("X-Real-Ip") != "" {
		ip = r.Header.Get("X-Real-Ip")
	} else {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	var browser = "Unknown"
	var os = "Unknown"
	userAgent := useragent.Parse(r.UserAgent())
	if userAgent != nil {
		browser = userAgent.Name
		if userAgent.Version.String() != "0.0.0" {
			browser += " " + userAgent.Version.String()
		}
		os = userAgent.OS
		if userAgent.OSVersion.String() != "0.0.0" {
			os += " " + userAgent.OSVersion.String()
		}
	}

	render.IPPage(r, w, types.ClientInfo{
		IPAddress: ip,
		Browser:   browser,
		OS:        os,
	})
}

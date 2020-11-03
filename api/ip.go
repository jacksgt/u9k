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

	userAgent := useragent.Parse(r.UserAgent())
	var browser string = userAgent.Name
	if userAgent.Version.String() != "0.0.0" {
		browser += " " + userAgent.Version.String()
	}
	var os string = userAgent.OS
	if userAgent.OSVersion.String() != "0.0.0" {
		os += " " + userAgent.OSVersion.String()
	}

	render.IPPage(r, w, types.ClientInfo{
		IPAddress: ip,
		Browser:   browser,
		OS:        os,
	})
}

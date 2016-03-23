package controllerhelpers

import (
	"net/http"
	"regexp"
)

var reIP = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`)

func GetIPAddr(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return reIP.FindStringSubmatch(ip)[1]
	}

	if reIP.MatchString(r.RemoteAddr) {
		return reIP.FindStringSubmatch(r.RemoteAddr)[1]
	}

	return "127.0.0.1"
}

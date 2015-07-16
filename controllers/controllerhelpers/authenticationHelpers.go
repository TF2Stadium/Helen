package controllerhelpers

import (
	"errors"
	"net/http"

	"github.com/TeamPlayTF/Server/config"
	"github.com/TeamPlayTF/Server/config/stores"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/sessions"
)

func buildFakeSocketRequest(cookiesObj *simplejson.Json) *http.Request {
	cookies, err := cookiesObj.Map()
	if err != nil {
		return &http.Request{}
	}

	str := ""

	first := true
	for k, v := range cookies {
		vStr, ok := v.(string)
		if !ok {
			continue
		}

		if !first {
			str += ";"
		}
		str += k + "=" + vStr
		first = false
	}

	if str == "" {
		return &http.Request{}
	}

	headers := http.Header{}
	headers.Add("Cookie", str)

	return &http.Request{Header: headers}
}

func AuthenticateSocket(socketid string, cookiesObj *simplejson.Json) {
	r := buildFakeSocketRequest(cookiesObj)
	s, _ := GetSessionHTTP(r)

	if _, ok := s.Values["playerid"]; ok {
		stores.SocketAuthStore[socketid] = s
	}
}

func DeauthenticateSocket(socketid string) {
	delete(stores.SocketAuthStore, socketid)
}

func IsLoggedInSocket(socketid string) bool {
	_, ok := stores.SocketAuthStore[socketid]
	return ok
}

func IsLoggedInHTTP(r *http.Request) bool {
	session, _ := stores.SessionStore.Get(r, config.Constants.SessionName)

	val, ok := session.Values["playerid"]
	return ok && val != ""
}

func GetSessionHTTP(r *http.Request) (*sessions.Session, error) {
	return stores.SessionStore.Get(r, config.Constants.SessionName)
}

func GetSessionSocket(socketid string) (*sessions.Session, error) {
	session, ok := stores.SocketAuthStore[socketid]

	if !ok {
		return nil, errors.New("No session associated with the socket")
	}
	return session, nil
}

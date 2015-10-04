// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"errors"
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
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

func AuthenticateSocket(socketid string, r *http.Request) error {
	s, _ := GetSessionHTTP(r)

	if _, ok := s.Values["id"]; ok {
		stores.SetSocketSession(socketid, s)
		return nil
	}

	return errors.New("Player isn't logged in")
}

func DeauthenticateSocket(socketid string) {
	stores.RemoveSocketSession(socketid)
}

func IsLoggedInSocket(socketid string) bool {
	_, ok := stores.GetStore(socketid)
	return ok
}

func IsLoggedInHTTP(r *http.Request) bool {
	session, _ := GetSessionHTTP(r)

	val, ok := session.Values["id"]
	return ok && val != ""
}

func GetSessionHTTP(r *http.Request) (*sessions.Session, error) {
	return stores.SessionStore.Get(r, config.Constants.SessionName)
}

func GetSessionSocket(socketid string) (*sessions.Session, error) {
	session, ok := stores.GetStore(socketid)

	if !ok {
		return nil, errors.New("No session associated with the socket")
	}
	return session, nil
}

func GetSteamId(socketid string) string {
	session, _ := GetSessionSocket(socketid)
	return session.Values["steam_id"].(string)
}

func GetPlayerSocket(socketid string) (*models.Player, error) {
	steamid := GetSteamId(socketid)
	return models.GetPlayerBySteamId(steamid)
}

func GetPlayerRole(socketid string) (authority.AuthRole, error) {
	session, err := GetSessionSocket(socketid)
	if err != nil {
		return 0, err
	}
	return session.Values["role"].(authority.AuthRole), nil
}

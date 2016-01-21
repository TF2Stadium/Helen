// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/gorilla/sessions"
)

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

func GetPlayerID(socketid string) uint {
	session, _ := GetSessionSocket(socketid)
	id, _ := strconv.ParseUint(session.Values["id"].(string), 10, 64)
	return uint(id)
}

func GetPlayerFromSocket(socketid string) *models.Player {
	id := GetPlayerID(socketid)
	player, _ := models.GetPlayerByID(id)
	return player
}

func GetPlayerRole(socketid string) (authority.AuthRole, error) {
	session, err := GetSessionSocket(socketid)
	if err != nil {
		return 0, err
	}
	return session.Values["role"].(authority.AuthRole), nil
}

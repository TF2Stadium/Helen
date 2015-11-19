// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"net/http"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

type FakeResponseWriter struct{}

func (f FakeResponseWriter) Header() http.Header {
	return http.Header{}
}
func (f FakeResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}
func (f FakeResponseWriter) WriteHeader(int) {}

func AdminChangeRole(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}
	var args struct {
		Steamid *string `json:"steamid"`
		Role    *string `json:"role"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	role, ok := helpers.RoleMap[*args.Role]
	if !ok || role == helpers.RoleAdmin {
		return helpers.NewTPError("Invalid role parameter", 0).Encode()
	}

	otherPlayer, err := models.GetPlayerBySteamId(*args.Steamid)
	if err != nil {
		return helpers.NewTPError("Player not found.", 0).Encode()
	}

	currPlayer, _ := chelpers.GetPlayerSocket(so.Id())

	models.LogAdminAction(currPlayer.ID, helpers.ActionChangeRole, otherPlayer.ID)

	// actual change happens
	otherPlayer.Role = role
	db.DB.Save(&otherPlayer)

	// rewrite session data. THIS WON'T WRITE A COOKIE SO IT ONLY WORKS WITH
	// STORES THAT STORE DATA IN COOKIES (AND NOT ONLY SESSION ID).
	session, sesserr := chelpers.GetSessionHTTP(so.Request())
	if sesserr == nil {
		session.Values["role"] = role
		session.Save(so.Request(), FakeResponseWriter{})
	}

	return chelpers.EmptySuccessJS
}

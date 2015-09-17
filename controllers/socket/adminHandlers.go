// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"net/http"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/googollee/go-socket.io"
	"reflect"
)

type FakeResponseWriter struct{}

func (f FakeResponseWriter) Header() http.Header {
	return http.Header{}
}
func (f FakeResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}
func (f FakeResponseWriter) WriteHeader(int) {}

var adminChangeRoleFilter = chelpers.FilterParams{
	Action:      helpers.ActionChangeRole,
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"steamid": chelpers.Param{Kind: reflect.String},
		"role":    chelpers.Param{Kind: reflect.String},
	},
}

//adminChangeRoleFilter,
//func(params map[string]interface{}) string {
//return ChangeRole(&so, params["role"].(string), params["steamid"].(string))
//})

func adminChangeRoleHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, adminChangeRoleFilter,

		func(params map[string]interface{}) string {
			roleString := params["role"].(string)
			steamid := params["steamid"].(string)
			role, ok := helpers.RoleMap[roleString]
			if !ok || role == helpers.RoleAdmin {
				bytes, _ := chelpers.BuildFailureJSON("Invalid role parameter", 0).Encode()
				return string(bytes)
			}

			otherPlayer, err := models.GetPlayerBySteamId(steamid)
			if err != nil {
				bytes, _ := chelpers.BuildFailureJSON("Player not found.", 0).Encode()
				return string(bytes)
			}

			currPlayer, _ := chelpers.GetPlayerSocket(so.Id())

			models.LogAdminAction(currPlayer.ID, helpers.ActionChangeRole, otherPlayer.ID)

			// actual change happens
			otherPlayer.Role = role
			db.DB.Save(&otherPlayer)

			// rewrite session data. THiS WON'T WRITE A COOKIE SO IT ONLY WORKS WITH
			// STORES THAT STORE DATA IN COOKIES (AND NOT ONLY SESSION ID).
			session, sesserr := chelpers.GetSessionHTTP(so.Request())
			if sesserr == nil {
				session.Values["role"] = role
				session.Save(so.Request(), FakeResponseWriter{})
			}

			return chelpers.BuildEmptySuccessString()
		})
}

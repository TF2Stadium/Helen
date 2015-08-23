package socket

import (
	"net/http"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
)

type FakeResponseWriter struct{}

func (f FakeResponseWriter) Header() http.Header {
	return http.Header{}
}
func (f FakeResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}
func (f FakeResponseWriter) WriteHeader(int) {}

var changeRoleParams = map[string]chelpers.Param{
	"steamid": chelpers.Param{Type: chelpers.PTypeString},
	"role":    chelpers.Param{Type: chelpers.PTypeString},
}

func ChangeRole(socket *socketio.Socket) func(string) string {
	socketid := (*socket).Id()
	return chelpers.AuthorizationFilter(socketid, helpers.ActionChangeRole,
		chelpers.JsonVerifiedFilter(changeRoleParams,
			func(js *simplejson.Json) string {
				roleString, _ := js.Get("role").String()
				role, ok := helpers.RoleMap[roleString]
				if !ok || role == helpers.RoleAdmin {
					bytes, _ := chelpers.BuildFailureJSON("Invalid role parameter.", 0).Encode()
					return string(bytes)
				}

				steamid, _ := js.Get("steamid").String()
				otherPlayer, err := models.GetPlayerBySteamId(steamid)
				if err != nil {
					bytes, _ := chelpers.BuildFailureJSON("Player not found.", 0).Encode()
					return string(bytes)
				}

				currPlayer, _ := chelpers.GetPlayerSocket(socketid)

				models.LogAdminAction(currPlayer.ID, helpers.ActionChangeRole, otherPlayer.ID)

				// actual change happens
				otherPlayer.Role = role
				db.DB.Save(&otherPlayer)

				//rewrite session data. THiS WON'T WRITE A COOKIE SO IT ONLY WORKS WITH
				// STORES THAT STORE DATA (AND NOT ONLY SESSION ID) IN COOKIES.
				session, sesserr := chelpers.GetSessionHTTP((*socket).Request())
				if sesserr == nil {
					session.Values["role"] = role
					session.Save((*socket).Request(), FakeResponseWriter{})
				}

				return chelpers.BuildEmptySuccessString()
			}))
}

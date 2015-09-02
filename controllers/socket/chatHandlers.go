package socket

import (
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
	"html"
	"reflect"
	"strconv"
	"time"
)

var chatSendFilter = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"message": chelpers.Param{Kind: reflect.String},
		"room":    chelpers.Param{Kind: reflect.Int},
	},
}

func chatSendHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, chatSendFilter,
		func(params map[string]interface{}) string {
			message := params["message"].(string)
			room := params["room"].(int)

			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			//Check if player has either joined, or is spectating lobby
			lobbyId, tperr := player.GetLobbyId()
			if room >= 0 {
				// if room is a lobby room
				if tperr != nil {
					bytes, _ := tperr.ErrorJSON().Encode()
					return string(bytes)
				} else if lobbyId != uint(room) && !player.IsSpectatingId(uint(room)) {
					bytes, _ := chelpers.BuildFailureJSON("Player is not in the lobby.", 5).Encode()
					return string(bytes)
				}
			} else {
				// else room is the lobby list room
				room = -1
			}

			t := time.Now()
			chatMessage := simplejson.New()
			// TODO send proper timestamps
			chatMessage.Set("timestamp", strconv.Itoa(t.Hour())+strconv.Itoa(t.Minute()))
			chatMessage.Set("message", html.EscapeString(message))
			chatMessage.Set("room", room)

			user := simplejson.New()
			user.Set("id", player.SteamId)
			user.Set("name", player.Name)

			chatMessage.Set("user", user)
			bytes, _ := chatMessage.Encode()
			broadcaster.SendMessageToRoom(strconv.Itoa(room), "chatReceive", string(bytes))

			resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(resp)
		})
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"math/rand"
	"reflect"
	"strconv"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
)

var debugLobbyFillFilter = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"id": chelpers.Param{Kind: reflect.Uint},
	},
}

func debugLobbyFillHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, debugLobbyFillFilter,
		func(param map[string]interface{}) string {
			id := param["id"].(uint)
			lobby, _ := models.GetLobbyById(id)
			var players []*models.Player

			for i := 1; i < models.TypePlayerCount[lobby.Type]*2; i++ {
				steamid := "DEBUG" + strconv.Itoa(int(rand.Int31()))

				player, _ := models.NewPlayer(steamid)
				player.Debug = true
				player.Save()
				players = append(players, player)
				lobby.AddPlayer(player, i)
			}

			lobby.State = models.LobbyStateReadyingUp
			lobby.Save()
			broadcaster.SendMessageToRoom(chelpers.GetLobbyRoom(lobby.ID), "lobbyReadyUp", "")

			for _, p := range players {
				lobby.ReadyPlayer(p)
			}
			lobby.State = models.LobbyStateInProgress
			lobby.Save()

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)

		})
}

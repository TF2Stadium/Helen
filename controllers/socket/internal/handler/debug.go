// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
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

func DebugLobbyFill(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, debugLobbyFillFilter,
		func(param map[string]interface{}) string {
			id := param["id"].(uint)
			lobby, _ := models.GetLobbyById(id)
			var players []*models.Player

			for i := 1; i < int(lobby.Type)*2; i++ {
				steamid := "DEBUG" + strconv.FormatUint(uint64(time.Now().Unix()), 10) + strconv.Itoa(i)

				player, _ := models.NewPlayer(steamid)
				player.Debug = true
				player.Save()
				players = append(players, player)
				lobby.AddPlayer(player, i)
			}

			lobby.State = models.LobbyStateReadyingUp
			lobby.Save()
			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobby.ID))
			broadcaster.SendMessageToRoom(room, "lobbyReadyUp", "")
			lobby.ReadyUpTimeoutCheck()
			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)

		})
}

var debugLobbyReadyFilter = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"id": chelpers.Param{Kind: reflect.Uint},
	},
}

func DebugLobbyReady(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, debugLobbyReadyFilter,
		func(param map[string]interface{}) string {
			id := param["id"].(uint)
			lobby, _ := models.GetLobbyById(id)

			var slots []models.LobbySlot
			db.DB.Where("lobby_id = ?", lobby.ID).Find(&slots)
			for _, slot := range slots {
				slot.Ready = true
				db.DB.Save(slot)
			}
			lobby.OnChange(true)

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)

		})
}

func DebugRequestAllLobbies(so socketio.Socket) func(string) string {
	return func(_ string) string {
		var lobbies []models.Lobby
		db.DB.Where("state <> ?", models.LobbyStateEnded).Find(&lobbies)
		list, err := models.DecorateLobbyListData(lobbies)

		if err != nil {
			helpers.Logger.Warning("Failed to send lobby list: %s", err.Error())
		} else {
			so.Emit("lobbyListData", list)
		}

		resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(resp)
	}
}

var debugRequestLobbyStartFilter = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"id": chelpers.Param{Kind: reflect.Uint},
	},
}

func DebugRequestLobbyStart(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, debugRequestLobbyStartFilter,
		func(params map[string]interface{}) string {
			lobby, _ := models.GetLobbyById(params["id"].(uint))
			bytes, _ := models.DecorateLobbyConnectJSON(lobby).Encode()
			room := fmt.Sprintf("%s_private", chelpers.GetLobbyRoom(lobby.ID))
			broadcaster.SendMessageToRoom(room,
				"lobbyStart", string(bytes))

			bytes, _ = chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})
}

var debugUpdateStatsFilter = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"id": chelpers.Param{Kind: reflect.Uint},
	},
}

func DebugUpdateStatsFilter(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, debugUpdateStatsFilter,
		func(params map[string]interface{}) string {
			lobby, err := models.GetLobbyById(params["id"].(uint))
			if err != nil {
				bytes, _ := err.ErrorJSON().Encode()
				return string(bytes)
			}
			lobby.UpdateStats()

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})
}

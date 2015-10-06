// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"crypto/rand"
	"encoding/base64"
	"reflect"
	"strconv"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
)

var lobbyCreateFilters = chelpers.FilterParams{
	Action:      authority.AuthAction(0),
	FilterLogin: true,

	Params: map[string]chelpers.Param{
		"mapName": chelpers.Param{Kind: reflect.String},

		"type": chelpers.Param{
			Kind: reflect.String,
			In:   []string{"highlander", "sixes"}},
		"league": chelpers.Param{
			Kind: reflect.String,
			In:   []string{"etf2l", "ugc"}},

		"server": chelpers.Param{Kind: reflect.String},

		"rconpwd":        chelpers.Param{Kind: reflect.String},
		"whitelist":      chelpers.Param{Kind: reflect.Uint},
		"mumbleRequired": chelpers.Param{Kind: reflect.Bool},
	},
}

func lobbyCreateHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbyCreateFilters,
		func(params map[string]interface{}) string {

			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			mapName := params["mapName"].(string)
			lobbytypestring := params["type"].(string)
			league := params["league"].(string)
			server := params["server"].(string)
			rconPwd := params["rconpwd"].(string)
			whitelist := int(params["whitelist"].(uint))
			//mumble := params["mumbleRequired"].(bool)

			var playermap = map[string]models.LobbyType{
				"sixes":      models.LobbyTypeSixes,
				"highlander": models.LobbyTypeHighlander,
			}

			lobbytype, ok := playermap[lobbytypestring]
			if !ok {
				bytes, _ := chelpers.BuildFailureJSON("Lobby type invalid.", -1).Encode()
				return string(bytes)
			}
			if !models.IsLeagueValid(league) {
				bytes, _ := chelpers.BuildFailureJSON("Invalid League Name", -1).Encode()
				return string(bytes)
			}

			randBytes := make([]byte, 6)
			rand.Read(randBytes)
			serverPwd := base64.URLEncoding.EncodeToString(randBytes)

			//TODO what if playermap[lobbytype] is nil?
			info := models.ServerRecord{
				Host:           server,
				RconPassword:   rconPwd,
				ServerPassword: serverPwd}
			err := models.VerifyInfo(info)
			if err != nil {
				return err.Error()
			}

			lob := models.NewLobby(mapName, lobbytype, league, info, whitelist)
			lob.CreatedBy = *player
			lob.Save()
			err = lob.SetupServer()

			if err != nil {
				bytes, _ := err.(*helpers.TPError).ErrorJSON().Encode()
				return string(bytes)
			}

			lob.State = models.LobbyStateWaiting
			lob.Save()
			lobby_id := simplejson.New()
			lobby_id.Set("id", lob.ID)
			bytes, _ := chelpers.BuildSuccessJSON(lobby_id).Encode()
			return string(bytes)
		})
}

var serverVerifyFilters = chelpers.FilterParams{
	Action:      authority.AuthAction(0),
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"server":  {Kind: reflect.String},
		"rconpwd": {Kind: reflect.String},
	},
}

func serverVerifyHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, serverVerifyFilters,
		func(params map[string]interface{}) string {
			info := models.ServerRecord{
				Host:         params["server"].(string),
				RconPassword: params["rconpwd"].(string),
			}
			err := models.VerifyInfo(info)
			if err != nil {
				bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
				return string(bytes)
			}

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)

		})
}

var lobbyCloseFilters = chelpers.FilterParams{
	Action:      authority.AuthAction(0),
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"id": chelpers.Param{Kind: reflect.Uint},
	},
}

func lobbyCloseHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbyCloseFilters,
		func(params map[string]interface{}) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			lobbyid := params["id"].(uint)

			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			if player.ID != lob.CreatedByID {
				bytes, _ := chelpers.BuildFailureJSON("Player not authorized to close lobby.", 1).Encode()
				return string(bytes)
			}

			if lob.State == models.LobbyStateEnded {
				bytes, _ := chelpers.BuildFailureJSON("Lobby already closed.", -1).Encode()
				return string(bytes)
			}

			helpers.LockRecord(lob.ID, lob)
			lob.Close(true)
			helpers.UnlockRecord(lob.ID, lob)
			chelpers.StopLogger(lobbyid)
			models.BroadcastLobbyList() // has to be done manually for now

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})

}

var lobbyJoinFilters = chelpers.FilterParams{
	Action:      authority.AuthAction(0),
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"id":    chelpers.Param{Kind: reflect.Uint},
		"class": chelpers.Param{Kind: reflect.String},
		"team":  chelpers.Param{Kind: reflect.String},
	},
}

func lobbyJoinHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbyJoinFilters,
		func(params map[string]interface{}) string {
			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lobbyid := params["id"].(uint)
			classString := params["class"].(string)
			teamString := params["team"].(string)

			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			slot, tperr := models.LobbyGetPlayerSlot(lob.Type, teamString, classString)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			helpers.LockRecord(lob.ID, lob)
			defer helpers.UnlockRecord(lob.ID, lob)
			tperr = lob.AddPlayer(player, slot)

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			chelpers.AfterLobbyJoin(so, lob, player)

			if lob.IsFull() {
				lob.State = models.LobbyStateReadyingUp
				lob.Save()
				lob.ReadyUpTimeoutCheck()
				broadcaster.SendMessageToRoom(
					chelpers.GetLobbyRoom(lob.ID),
					"lobbyReadyUp", "")
				models.BroadcastLobbyList()
			}

			models.BroadcastLobbyToUser(lob, player.SteamId)
			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})
}

var lobbySpectatorJoinFilters = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"id": chelpers.Param{Kind: reflect.Uint},
	},
}

func lobbySpectatorJoinHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbySpectatorJoinFilters,
		func(params map[string]interface{}) string {

			lobbyid := params["id"].(uint)

			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}
			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}
			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()

			helpers.LockRecord(lob.ID, lob)
			tperr = lob.AddSpectator(player)
			helpers.UnlockRecord(lob.ID, lob)

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			chelpers.AfterLobbyJoin(so, lob, player)
			models.BroadcastLobbyToUser(lob, player.SteamId)
			return string(bytes)
		})
}

var lobbyKickFilters = chelpers.FilterParams{
	Action:      authority.AuthAction(0),
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"id":      chelpers.Param{Kind: reflect.Uint},
		"steamid": chelpers.Param{Kind: reflect.String, Default: ""},
		"ban":     chelpers.Param{Kind: reflect.Bool, Default: false},
	},
}

func lobbyKickHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbyKickFilters,
		func(params map[string]interface{}) string {
			steamid := params["steamid"].(string)
			ban := params["ban"].(bool)
			lobbyid := params["id"].(uint)
			self := false

			// TODO check authorization, currently can kick anyone
			if steamid == "" || steamid == chelpers.GetSteamId(so.Id()) {
				self = true
				steamid = chelpers.GetSteamId(so.Id())
			}

			player, tperr := models.GetPlayerBySteamId(steamid)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := chelpers.BuildFailureJSON(tperr.Error(), -1).Encode()
				return string(bytes)
			}

			if !self && lob.CreatedByID != player.ID {
				// TODO proper authorization checks
				bytes, _ := chelpers.BuildFailureJSON("Not authorized to remove players", 1).Encode()
				return string(bytes)
			}

			_, err := lob.GetPlayerSlot(player)
			helpers.LockRecord(lob.ID, lob)
			defer helpers.UnlockRecord(lob.ID, lob)

			if err == nil {
				lob.RemovePlayer(player)
			} else if player.IsSpectatingId(lob.ID) {
				lob.RemoveSpectator(player)
			} else {
				bytes, _ := chelpers.BuildFailureJSON("Player neither playing nor spectating", 2).Encode()
				return string(bytes)
			}

			if ban {
				lob.BanPlayer(player)
			}

			chelpers.AfterLobbyLeave(so, lob, player)
			so.Emit("lobbyData", "{}")
			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})
}

var playerReadyFilter = chelpers.FilterParams{
	Action:      authority.AuthAction(0),
	FilterLogin: true,
}

func playerReadyHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, playerReadyFilter,
		func(_ map[string]interface{}) string {
			steamid := chelpers.GetSteamId(so.Id())
			player, tperr := models.GetPlayerBySteamId(steamid)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lobbyid, tperr := player.GetLobbyId()
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lobby, tperr := models.GetLobbyById(lobbyid)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			if lobby.State != models.LobbyStateReadyingUp {
				bytes, _ := helpers.NewTPError("Lobby hasn't been filled up yet.", 4).ErrorJSON().Encode()
				return string(bytes)
			}

			helpers.LockRecord(lobby.ID, lobby)
			tperr = lobby.ReadyPlayer(player)
			defer helpers.UnlockRecord(lobby.ID, lobby)

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			if lobby.IsEveryoneReady() {
				lobby.State = models.LobbyStateInProgress
				lobby.Save()
				bytes, _ := models.DecorateLobbyConnectJSON(lobby).Encode()
				broadcaster.SendMessageToRoom(strconv.FormatUint(uint64(lobby.ID), 10),
					"lobbyStart", string(bytes))
				models.BroadcastLobbyList()
			}

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})
}

var playerUnreadyFilter = chelpers.FilterParams{
	Action:      authority.AuthAction(0),
	FilterLogin: true,
}

func playerUnreadyHandler(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, playerUnreadyFilter,
		func(_ map[string]interface{}) string {
			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lobbyid, tperr := player.GetLobbyId()
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lobby, tperr := models.GetLobbyById(lobbyid)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			helpers.LockRecord(lobby.ID, lobby)
			tperr = lobby.UnreadyPlayer(player)
			helpers.UnlockRecord(lobby.ID, lobby)

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})
}

func requestLobbyListDataHandler(so socketio.Socket) func(string) string {
	return func(s string) string {
		var lobbies []models.Lobby
		db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies)
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

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"reflect"

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
		"map": chelpers.Param{Kind: reflect.String},

		"type": chelpers.Param{
			Kind: reflect.String,
			In:   []string{"highlander", "sixes", "debug"}},
		"league": chelpers.Param{
			Kind: reflect.String,
			In:   []string{"etf2l", "ugc", "esea", "asiafortress", "ozfortress"}},

		"server": chelpers.Param{Kind: reflect.String},

		"rconpwd":        chelpers.Param{Kind: reflect.String},
		"whitelistID":    chelpers.Param{Kind: reflect.Uint},
		"mumbleRequired": chelpers.Param{Kind: reflect.Bool},
	},
}

func LobbyCreate(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbyCreateFilters,
		func(params map[string]interface{}) string {

			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			mapName := params["map"].(string)
			lobbytypestring := params["type"].(string)
			league := params["league"].(string)
			server := params["server"].(string)
			rconPwd := params["rconpwd"].(string)
			whitelist := int(params["whitelistID"].(uint))
			mumble := params["mumbleRequired"].(bool)

			var playermap = map[string]models.LobbyType{
				"debug":      models.LobbyTypeDebug,
				"sixes":      models.LobbyTypeSixes,
				"highlander": models.LobbyTypeHighlander,
			}

			lobbytype, _ := playermap[lobbytypestring]

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

			lob := models.NewLobby(mapName, lobbytype, league, info, whitelist, mumble)
			lob.CreatedBySteamID = player.SteamId
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

func ServerVerify(so socketio.Socket) func(string) string {
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

func LobbyClose(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbyCloseFilters,
		func(params map[string]interface{}) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			lobbyid := params["id"].(uint)

			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			if player.SteamId != lob.CreatedBySteamID && player.Role != helpers.RoleAdmin {
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

func LobbyJoin(so socketio.Socket) func(string) string {
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

			//Check if player is in the same lobby
			var sameLobby bool
			if id, err := player.GetLobbyId(); err == nil && id == lobbyid {
				sameLobby = true
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

			if !sameLobby {
				chelpers.AfterLobbyJoin(so, lob, player)
			}

			if lob.IsFull() {
				lob.State = models.LobbyStateReadyingUp
				lob.Save()
				lob.ReadyUpTimeoutCheck()
				room := fmt.Sprintf("%s_private",
					chelpers.GetLobbyRoom(lob.ID))
				broadcaster.SendMessageToRoom(room, "lobbyReadyUp",
					`{"timeout":30}`)
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

func LobbySpectatorJoin(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbySpectatorJoinFilters,
		func(params map[string]interface{}) string {

			lobbyid := params["id"].(uint)

			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			var lob *models.Lobby
			lob, tperr = models.GetLobbyById(uint(lobbyid))

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			if id, _ := player.GetLobbyId(); id != lobbyid {
				helpers.LockRecord(lob.ID, lob)
				tperr = lob.AddSpectator(player)
				helpers.UnlockRecord(lob.ID, lob)
			}

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			chelpers.AfterLobbySpec(so, lob)
			models.BroadcastLobbyToUser(lob, player.SteamId)
			return string(bytes)
		})
}

var lobbyNoLoginSpectatorJoinFilters = chelpers.FilterParams{
	Params: map[string]chelpers.Param{
		"id": chelpers.Param{Kind: reflect.Uint},
	},
}

func LobbyNoLoginSpectatorJoin(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbyNoLoginSpectatorJoinFilters,
		func(params map[string]interface{}) string {
			id := params["id"].(uint)

			lobby, err := models.GetLobbyById(id)
			if err != nil {
				bytes, _ := err.ErrorJSON().Encode()
				return string(bytes)
			}

			chelpers.AfterLobbySpec(so, lobby)
			bytes, _ := models.DecorateLobbyDataJSON(lobby, true).Encode()
			so.Emit("lobbyData", string(bytes))

			bytes, _ = chelpers.BuildSuccessJSON(simplejson.New()).Encode()
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

func LobbyKick(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, lobbyKickFilters,
		func(params map[string]interface{}) string {
			steamid := params["steamid"].(string)
			ban := params["ban"].(bool)
			lobbyid := params["id"].(uint)
			var self bool

			selfSteamid := chelpers.GetSteamId(so.Id())
			// TODO check authorization, currently can kick anyone
			if steamid == "" {
				self = true
				steamid = selfSteamid
			}

			//player to kick
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

			if !self && selfSteamid != lob.CreatedBySteamID {
				// TODO proper authorization checks
				bytes, _ := chelpers.BuildFailureJSON(
					"Not authorized to remove players", 1).Encode()
				return string(bytes)
			}

			_, err := lob.GetPlayerSlot(player)
			helpers.LockRecord(lob.ID, lob)
			defer helpers.UnlockRecord(lob.ID, lob)

			var spec bool
			if err == nil {
				lob.RemovePlayer(player)
			} else if player.IsSpectatingId(lob.ID) {
				spec = true
				lob.RemoveSpectator(player, true)
			} else {
				bytes, _ := chelpers.BuildFailureJSON("Player neither playing nor spectating", 2).Encode()
				return string(bytes)
			}

			if ban {
				lob.BanPlayer(player)
			}

			if !spec {
				chelpers.AfterLobbyLeave(so, lob, player)
			} else {
				chelpers.AfterLobbySpecLeave(so, lob)
			}

			if !self {
				broadcaster.SendMessage(steamid, "sendNotification",
					fmt.Sprintf("You have been removed from Lobby #%d", lobbyid))

			}

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})
}

var playerReadyFilter = chelpers.FilterParams{
	Action:      authority.AuthAction(0),
	FilterLogin: true,
}

func PlayerReady(so socketio.Socket) func(string) string {
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
				room := fmt.Sprintf("%s_private",
					chelpers.GetLobbyRoom(lobby.ID))
				broadcaster.SendMessageToRoom(room,
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

func PlayerNotReady(so socketio.Socket) func(string) string {
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

			if lobby.State != models.LobbyStateReadyingUp {
				bytes, _ := helpers.NewTPError("Lobby hasn't been filled up yet.", 4).ErrorJSON().Encode()
				return string(bytes)
			}

			helpers.LockRecord(lobby.ID, lobby)
			tperr = lobby.UnreadyPlayer(player)
			lobby.RemovePlayer(player)
			helpers.UnlockRecord(lobby.ID, lobby)

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lobby.UnreadyAllPlayers()

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})
}

func RequestLobbyListData(so socketio.Socket) func(string) string {
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

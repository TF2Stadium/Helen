// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"fmt"
	"strconv"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
)

var BanTypeList = []string{"join", "create", "chat", "full"}

var BanTypeMap = map[string]models.PlayerBanType{
	"join":   models.PlayerBanJoin,
	"create": models.PlayerBanCreate,
	"chat":   models.PlayerBanChat,
	"full":   models.PlayerBanFull,
}

func AfterLobbyJoin(so socketio.Socket, lobby *models.Lobby, player *models.Player) {
	so.Join(fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID)))
}

func AfterLobbyLeave(so socketio.Socket, lobby *models.Lobby, player *models.Player) {
	so.Leave(fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID)))
	so.Leave(fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
}

func AfterLobbySpec(so socketio.Socket, lobby *models.Lobby) {
	so.Join(fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
	BroadcastScrollback(so, lobby.ID)
}

func AfterLobbySpecLeave(so socketio.Socket, lobby *models.Lobby) {
	so.Leave(fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
}

func AfterConnect(so socketio.Socket) {
	so.Join(fmt.Sprintf("%s_public", config.Constants.GlobalChatRoom)) //room for global chat

	var lobbies []models.Lobby
	err := db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies).Error
	if err != nil {
		helpers.Logger.Critical("%s", err.Error())
		return
	}

	list, err := models.DecorateLobbyListData(lobbies)
	if err != nil {
		helpers.Logger.Critical("Failed to send lobby list: %s", err.Error())
		return
	}

	so.Emit("lobbyListData", list)
	BroadcastScrollback(so, 0)
}

func AfterConnectLoggedIn(so socketio.Socket, player *models.Player) {
	lobbyIdPlaying, err := player.GetLobbyId()
	if err == nil {
		lobby, _ := models.GetLobbyById(lobbyIdPlaying)
		AfterLobbyJoin(so, lobby, player)
		AfterLobbySpec(so, lobby)
		models.BroadcastLobbyToUser(lobby, GetSteamId(so.Id()))
		slot := &models.LobbySlot{}
		err := db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(slot).Error
		if err == nil {
			if lobby.State == models.LobbyStateInProgress && !slot.InGame {
				bytes, _ := models.DecorateLobbyConnectJSON(lobby).Encode()
				broadcaster.SendMessage(player.SteamId, "lobbyStart", string(bytes))
			} else if lobby.State == models.LobbyStateReadyingUp && !slot.Ready {
				left := simplejson.New()
				left.Set("timeout", lobby.ReadyUpTimeLeft())
				bytes, _ := left.Encode()
				broadcaster.SendMessage(player.SteamId, "lobbyReadyUp", string(bytes))
			}
		}
	}

	settings, err2 := player.GetSettings()
	if err2 == nil {
		json := models.DecoratePlayerSettingsJson(settings)
		bytes, _ := json.Encode()
		broadcaster.SendMessage(player.SteamId, "playerSettings", string(bytes))
	}

	profilePlayer, err3 := models.GetPlayerWithStats(player.SteamId)
	if err3 == nil {
		json := models.DecoratePlayerProfileJson(profilePlayer)
		bytes, _ := json.Encode()
		broadcaster.SendMessage(player.SteamId, "playerProfile", string(bytes))
	}

}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}

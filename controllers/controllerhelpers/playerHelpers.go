// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
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
	so.Join(GetLobbyRoom(lobby.ID))
}

func AfterLobbyLeave(so socketio.Socket, lobby *models.Lobby, player *models.Player) {
	so.Leave(GetLobbyRoom(lobby.ID))
}

func AfterConnect(so socketio.Socket) {
	so.Join(config.Constants.GlobalChatRoom) //room for global chat

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
}

func AfterConnectLoggedIn(so socketio.Socket, player *models.Player) {
	lobbyIdPlaying, err := player.GetLobbyId()
	if err == nil {
		so.Join(GetLobbyRoom(lobbyIdPlaying))
		lobby, _ := models.GetLobbyById(lobbyIdPlaying)
		models.BroadcastLobbyToUser(lobby, GetSteamId(so.Id()))
		slot := models.LobbySlot{}
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

	lobbyIdsSpectating, err2 := player.GetSpectatingIds()
	if err2 == nil {
		for _, id := range lobbyIdsSpectating {
			so.Join(GetLobbyRoom(id))
			lobby, _ := models.GetLobbyById(id)
			models.BroadcastLobbyToUser(lobby, GetSteamId(so.Id()))
		}
	}
}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

var BanTypeList = []string{"join", "create", "chat", "full"}

var BanTypeMap = map[string]models.PlayerBanType{
	"join":   models.PlayerBanJoin,
	"create": models.PlayerBanCreate,
	"chat":   models.PlayerBanChat,
	"full":   models.PlayerBanFull,
}

func AfterLobbyJoin(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby, player *models.Player) {
	room := fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID))
	server.AddClient(so, room)

	bytes, _ := json.Marshal(models.DecorateLobbyData(lobby, false))
	broadcaster.SendMessage(player.SteamId, "lobbyJoined", string(bytes))
}

func AfterLobbyLeave(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby, player *models.Player) {
	server.RemoveClient(so.Id(), fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID)))
	server.RemoveClient(so.Id(), fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))

	bytes, _ := json.Marshal(models.DecorateLobbyData(lobby, true))
	broadcaster.SendMessageToRoom(player.SteamId, "lobbyData", string(bytes))

	bytes, _ = json.Marshal(models.DecorateLobbyLeave(lobby))
	broadcaster.SendMessage(player.SteamId, "lobbyLeft", string(bytes))
}

func AfterLobbySpec(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby) {
	server.AddClient(so, fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
	BroadcastScrollback(so, lobby.ID)
}

func AfterLobbySpecLeave(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby) {
	server.RemoveClient(so.Id(), fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
}

func AfterConnect(server *wsevent.Server, so *wsevent.Client) {
	server.AddClient(so, fmt.Sprintf("%s_public", config.Constants.GlobalChatRoom)) //room for global chat

	var lobbies []models.Lobby
	err := db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies).Error
	if err != nil {
		helpers.Logger.Critical("%s", err.Error())
		return
	}

	bytes, _ := json.Marshal(models.DecorateLobbyListData(lobbies))
	if err != nil {
		helpers.Logger.Critical("Failed to send lobby list: %s", err.Error())
		return
	}

	so.EmitJSON(helpers.NewRequest("lobbyListData", string(bytes)))
	BroadcastScrollback(so, 0)
	bytes, _ = json.Marshal(helpers.NewRequestFromObj("subListData", models.GetSubList()))
	so.EmitJSON(helpers.NewRequest("subListData", string(bytes)))
}

func AfterConnectLoggedIn(server *wsevent.Server, so *wsevent.Client, player *models.Player) {
	lobbyIdPlaying, err := player.GetLobbyId()
	if err == nil {
		lobby, _ := models.GetLobbyByIdServer(lobbyIdPlaying)
		AfterLobbyJoin(server, so, lobby, player)
		AfterLobbySpec(server, so, lobby)
		models.BroadcastLobbyToUser(lobby, GetSteamId(so.Id()))
		slot := &models.LobbySlot{}
		err := db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(slot).Error
		if err == nil {
			if lobby.State == models.LobbyStateInProgress && !models.IsPlayerInServer(player.SteamId) {
				bytes, _ := json.Marshal(models.DecorateLobbyConnect(lobby))
				broadcaster.SendMessage(player.SteamId, "lobbyStart", string(bytes))
			} else if lobby.State == models.LobbyStateReadyingUp && !slot.Ready {
				data := struct {
					Timeout int64 `json:"timeout"`
				}{lobby.ReadyUpTimeLeft()}

				bytes, _ := json.Marshal(data)
				broadcaster.SendMessage(player.SteamId, "lobbyReadyUp", string(bytes))
			}
		}
	}

	settings, err2 := player.GetSettings()
	if err2 == nil {
		bytes, _ := json.Marshal(models.DecoratePlayerSettingsJson(settings))
		broadcaster.SendMessage(player.SteamId, "playerSettings", string(bytes))
	}

	profilePlayer, err3 := models.GetPlayerWithStats(player.SteamId)
	if err3 == nil {
		bytes, _ := json.Marshal(models.DecoratePlayerProfileJson(profilePlayer))
		broadcaster.SendMessage(player.SteamId, "playerProfile", string(bytes))
	}

}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}

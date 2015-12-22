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

	broadcaster.SendMessage(player.SteamID, "lobbyJoined", models.DecorateLobbyData(lobby, false))
}

func AfterLobbyLeave(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby, player *models.Player) {
	//pub := fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID))

	// bytes, _ := json.Marshal(models.DecorateLobbyData(lobby, true))
	// broadcaster.SendMessageToRoom(pub, "lobbyData", string(bytes))

	broadcaster.SendMessage(player.SteamID, "lobbyLeft", models.DecorateLobbyLeave(lobby))

	server.RemoveClient(so.Id(), fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID)))
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

	so.EmitJSON(helpers.NewRequest("lobbyListData", models.DecorateLobbyListData(lobbies)))
	BroadcastScrollback(so, 0)
	so.EmitJSON(helpers.NewRequest("subListData", models.GetAllSubs()))
}

func AfterConnectLoggedIn(server *wsevent.Server, so *wsevent.Client, player *models.Player) {
	lobbyIdPlaying, err := player.GetLobbyID()
	if err == nil {
		lobby, _ := models.GetLobbyByIdServer(lobbyIdPlaying)
		AfterLobbyJoin(server, so, lobby, player)
		AfterLobbySpec(server, so, lobby)
		models.BroadcastLobbyToUser(lobby, GetSteamId(so.Id()))
		slot := &models.LobbySlot{}
		err := db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(slot).Error
		if err == nil {
			if lobby.State == models.LobbyStateInProgress && !models.IsPlayerInServer(player.SteamID) {
				_, class, _ := models.LobbyGetSlotInfoString(lobby.Type, slot.Slot)
				broadcaster.SendMessage(player.SteamID, "lobbyStart", models.DecorateLobbyConnect(lobby, player.Name, class))
			} else if lobby.State == models.LobbyStateReadyingUp && !slot.Ready {
				data := struct {
					Timeout int64 `json:"timeout"`
				}{lobby.ReadyUpTimeLeft()}

				broadcaster.SendMessage(player.SteamID, "lobbyReadyUp", data)
			}
		}
	}

	settings, err2 := player.GetSettings()
	if err2 == nil {
		broadcaster.SendMessage(player.SteamID, "playerSettings", models.DecoratePlayerSettingsJson(settings))
	}

	profilePlayer, err3 := models.GetPlayerWithStats(player.SteamID)
	if err3 == nil {
		broadcaster.SendMessage(player.SteamID, "playerProfile", models.DecoratePlayerProfileJson(profilePlayer))
	}

}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}

//Not really broadcast, since it sends each client a different LobbyStart JSON
func BroadcastLobbyStart(lobby *models.Lobby) {
	var slots []*models.LobbySlot

	db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).Find(&slots)

	for _, slot := range slots {
		var player models.Player
		db.DB.First(&player, slot.PlayerID)

		_, class, _ := models.LobbyGetSlotInfoString(lobby.Type, slot.Slot)
		broadcaster.SendMessage(player.SteamID, "lobbyStart", models.DecorateLobbyConnect(lobby, player.Name, class))
	}
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package hooks

import (
	"fmt"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

func AfterConnect(server *wsevent.Server, so *wsevent.Client) {
	server.AddClient(so, fmt.Sprintf("%s_public", config.Constants.GlobalChatRoom)) //room for global chat

	var lobbies []models.Lobby
	err := db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies).Error
	if err != nil {
		helpers.Logger.Critical("%s", err.Error())
		return
	}

	so.EmitJSON(helpers.NewRequest("lobbyListData", models.DecorateLobbyListData(lobbies)))
	chelpers.BroadcastScrollback(so, 0)
	so.EmitJSON(helpers.NewRequest("subListData", models.GetAllSubs()))
}

func AfterConnectLoggedIn(server *wsevent.Server, so *wsevent.Client, player *models.Player) {
	if time.Since(player.UpdatedAt) >= time.Hour*1 {
		player.UpdatePlayerInfo()
		player.Save()
	}

	lobbyID, err := player.GetLobbyID(false)
	if err == nil {
		lobby, _ := models.GetLobbyByIDServer(lobbyID)
		AfterLobbyJoin(server, so, lobby, player)
		AfterLobbySpec(server, so, lobby)
		models.BroadcastLobbyToUser(lobby, chelpers.GetSteamId(so.Id()))
		slot := &models.LobbySlot{}
		err := db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(slot).Error
		if err == nil {
			if lobby.State == models.LobbyStateInProgress && slot.InGame {
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

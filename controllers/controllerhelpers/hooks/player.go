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
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

func AfterConnect(server *wsevent.Server, so *wsevent.Client) {
	server.Join(so, fmt.Sprintf("%s_public", config.GlobalChatRoom)) //room for global chat

	so.EmitJSON(helpers.NewRequest("lobbyListData", models.DecorateLobbyListData(models.GetWaitingLobbies())))
	chelpers.BroadcastScrollback(so, 0)
	so.EmitJSON(helpers.NewRequest("subListData", models.DecorateSubstituteList()))
}

func AfterConnectLoggedIn(so *wsevent.Client, player *models.Player) {
	if time.Since(player.UpdatedAt) >= time.Hour*1 {
		player.UpdatePlayerInfo()
		player.Save()
	}

	lobbyID, err := player.GetLobbyID(false)
	if err == nil {
		lobby, _ := models.GetLobbyByIDServer(lobbyID)
		AfterLobbyJoin(so, lobby, player)
		AfterLobbySpec(socket.AuthServer, so, lobby)
		models.BroadcastLobbyToUser(lobby, so.Token.Claims["steam_id"].(string))

		slot := &models.LobbySlot{}
		err := db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(slot).Error

		if err == nil {
			if lobby.State == models.LobbyStateInProgress {
				broadcaster.SendMessage(player.SteamID, "lobbyStart", models.DecorateLobbyConnect(lobby, player.Name, slot.Slot))
			} else if lobby.State == models.LobbyStateReadyingUp && !slot.Ready {
				data := struct {
					Timeout int64 `json:"timeout"`
				}{lobby.ReadyUpTimeLeft()}

				broadcaster.SendMessage(player.SteamID, "lobbyReadyUp", data)
			}
		}
	}

	settings := player.Settings
	if settings != nil {
		broadcaster.SendMessage(player.SteamID, "playerSettings", settings)
	}
	profilePlayer, err3 := models.GetPlayerWithStats(player.SteamID)
	if err3 == nil {
		broadcaster.SendMessage(player.SteamID, "playerProfile", models.DecoratePlayerProfileJson(profilePlayer))
	}

}

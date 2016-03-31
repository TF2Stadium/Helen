// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package hooks

import (
	"time"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

func AfterConnect(server *wsevent.Server, so *wsevent.Client) {
	server.Join(so, "0_public") //room for global chat

	so.EmitJSON(helpers.NewRequest("lobbyListData", lobby.DecorateLobbyListData(lobby.GetWaitingLobbies())))
	chelpers.BroadcastScrollback(so, 0)
	so.EmitJSON(helpers.NewRequest("subListData", lobby.DecorateSubstituteList()))
}

func AfterConnectLoggedIn(so *wsevent.Client, player *player.Player) {
	if time.Since(player.ProfileUpdatedAt) >= 30*time.Minute {
		player.UpdatePlayerInfo()
	}

	lobbyID, err := player.GetLobbyID(false)
	if err == nil {
		lob, _ := lobby.GetLobbyByIDServer(lobbyID)
		AfterLobbyJoin(so, lob, player)
		AfterLobbySpec(socket.AuthServer, so, player, lob)
		lobby.BroadcastLobbyToUser(lob, so.Token.Claims["steam_id"].(string))

		slot := &lobby.LobbySlot{}
		err := db.DB.Where("lobby_id = ? AND player_id = ?", lob.ID, player.ID).First(slot).Error

		if err == nil {
			if lob.State == lobby.InProgress {
				broadcaster.SendMessage(player.SteamID, "lobbyStart", lobby.DecorateLobbyConnect(lob, player, slot.Slot))
			} else if lob.State == lobby.ReadyingUp && !slot.Ready {
				data := struct {
					Timeout int64 `json:"timeout"`
				}{lob.ReadyUpTimeLeft()}

				broadcaster.SendMessage(player.SteamID, "lobbyReadyUp", data)
			}
		}
	}

	settings := player.Settings
	if settings != nil {
		broadcaster.SendMessage(player.SteamID, "playerSettings", settings)
	} else {
		broadcaster.SendMessage(player.SteamID, "playerSettings", map[string]string{})
	}

	player.SetPlayerProfile()
	broadcaster.SendMessage(player.SteamID, "playerProfile", player)
}

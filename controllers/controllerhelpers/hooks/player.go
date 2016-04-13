// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package hooks

import (
	"time"

	"github.com/TF2Stadium/Helen/config"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

func AfterConnect(server *wsevent.Server, so *wsevent.Client) {
	server.Join(so, "0_public") //room for global chat

	so.EmitJSON(helpers.NewRequest("lobbyListData", lobby.DecorateLobbyListData(lobby.GetWaitingLobbies(), false)))
	chelpers.BroadcastScrollback(so, 0)
	so.EmitJSON(helpers.NewRequest("subListData", lobby.DecorateSubstituteList()))
}

var emptyMap = make(map[string]string)

func AfterConnectLoggedIn(so *wsevent.Client, player *player.Player) {
	sessions.AddSocket(player.SteamID, so)

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
				so.EmitJSON(helpers.NewRequest("lobbyStart", lobby.DecorateLobbyConnect(lob, player, slot.Slot)))
			} else if lob.State == lobby.ReadyingUp && !slot.Ready {
				data := struct {
					Timeout int64 `json:"timeout"`
				}{lob.ReadyUpTimeLeft()}

				so.EmitJSON(helpers.NewRequest("lobbyReadyUp", data))
			}
		}
	}

	if player.Settings != nil {
		so.EmitJSON(helpers.NewRequest("playerSettings", player.Settings))
	} else {
		so.EmitJSON(helpers.NewRequest("playerSettings", emptyMap))
	}

	player.SetPlayerProfile()
	so.EmitJSON(helpers.NewRequest("playerProfile", player))
	so.EmitJSON(helpers.NewRequest("mumbleInfo", struct {
		Address  string `json:"address"`
		Password string `json:"password"`
	}{config.Constants.MumbleAddr, player.MumbleAuthkey}))
}

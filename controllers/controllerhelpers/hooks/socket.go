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
	"github.com/TF2Stadium/Helen/models"
)

//OnDisconnect is connected when a player with a given socketID disconnects
func OnDisconnect(socketID string) {
	defer chelpers.DeauthenticateSocket(socketID)
	if chelpers.IsLoggedInSocket(socketID) {
		steamid := chelpers.GetSteamId(socketID)
		broadcaster.RemoveSocket(socketID, steamid)
		player, tperr := models.GetPlayerBySteamID(steamid)
		if tperr != nil || player == nil {
			helpers.Logger.Error(tperr.Error())
			return
		}

		ids, tperr := player.GetSpectatingIds()
		if tperr != nil {
			helpers.Logger.Error(tperr.Error())
			return
		}

		for _, id := range ids {
			//if this _specific_ socket is spectating this lobby, remove them from it
			//player might be spectating other lobbies in another tab, but we don't care
			if broadcaster.IsSpectating(socketID, id) {
				lobby, _ := models.GetLobbyByID(id)
				err := lobby.RemoveSpectator(player, true)
				if err != nil {
					helpers.Logger.Error(err.Error())
					continue
				}
				broadcaster.RemoveSpectator(socketID)
				//helpers.Logger.Debug("removing %s from %d", player.SteamId, id)
			}
		}

		id, _ := player.GetLobbyID(true)
		//if player is in a waiting lobby, and hasn't connected for > 30 seconds,
		//remove him from it. Here, connected = player isn't connected from any tab/window
		if id != 0 && broadcaster.ConnectedSockets(player.SteamID) == 0 {
			time.AfterFunc(time.Second*30, func() {
				if !broadcaster.IsConnected(player.SteamID) {
					//player may have changed lobbies during this time
					//fetch lobby ID again
					id, err := player.GetLobbyID(true)
					if err != nil {
						return
					}

					lobby := &models.Lobby{}
					db.DB.First(lobby, id)
					lobby.RemovePlayer(player)
				}
			})
		}
	}

}

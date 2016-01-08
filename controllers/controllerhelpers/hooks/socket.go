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
		broadcaster.RemoveSocket(steamid)
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
			lobby, _ := models.GetLobbyByID(id)
			err := lobby.RemoveSpectator(player, true)
			if err != nil {
				helpers.Logger.Error(err.Error())
				continue
			}
			//helpers.Logger.Debug("removing %s from %d", player.SteamId, id)
		}

		time.AfterFunc(time.Second*30, func() {
			if !broadcaster.IsConnected(player.SteamID) {
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

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package hooks

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/internal/pprof"
	"github.com/TF2Stadium/Helen/models"
	"github.com/dgrijalva/jwt-go"
)

//OnDisconnect is connected when a player with a given socketID disconnects
func OnDisconnect(socketID string, token *jwt.Token) {
	pprof.Clients.Add(-1)
	if token != nil { //player was logged in
		steamid := token.Claims["steam_id"].(string)
		sessions.RemoveSocket(socketID, steamid)
		player, tperr := models.GetPlayerBySteamID(steamid)
		if tperr != nil || player == nil {
			logrus.Error(tperr.Error())
			return
		}

		ids, tperr := player.GetSpectatingIds()
		if tperr != nil {
			logrus.Error(tperr.Error())
			return
		}

		for _, id := range ids {
			//if this _specific_ socket is spectating this lobby, remove them from it
			//player might be spectating other lobbies in another tab, but we don't care
			if sessions.IsSpectating(socketID, id) {
				lobby, _ := models.GetLobbyByID(id)
				err := lobby.RemoveSpectator(player, true)
				if err != nil {
					logrus.Error(err.Error())
					continue
				}
				sessions.RemoveSpectator(socketID)
				//logrus.Debug("removing %s from %d", player.SteamId, id)
			}
		}

		id, _ := player.GetLobbyID(true)
		//if player is in a waiting lobby, and hasn't connected for > 30 seconds,
		//remove him from it. Here, connected = player isn't connected from any tab/window
		if id != 0 && sessions.ConnectedSockets(player.SteamID) == 0 {
			time.AfterFunc(time.Second*30, func() {
				if !sessions.IsConnected(player.SteamID) {
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

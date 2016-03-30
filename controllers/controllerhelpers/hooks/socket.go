// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package hooks

import (
	"time"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/dgrijalva/jwt-go"
)

//OnDisconnect is connected when a player with a given socketID disconnects
func OnDisconnect(socketID string, token *jwt.Token) {
	if token != nil { //player was logged in
		player := chelpers.GetPlayer(token)
		if player == nil {
			return
		}

		sessions.RemoveSocket(socketID, player.SteamID)
		id, _ := sessions.GetSpectating(socketID)
		if id != 0 {
			lob, _ := lobby.GetLobbyByID(id)
			lob.RemoveSpectator(player, true)
		}

		id, _ = player.GetLobbyID(true)
		//if player is in a waiting lobby, and hasn't connected for > 30 seconds,
		//remove him from it. Here, connected = player isn't connected from any tab/window
		if id != 0 && sessions.ConnectedSockets(player.SteamID) == 0 {
			sessions.AfterDisconnectedFunc(player.SteamID, time.Second*30, func() {
				lob, _ := lobby.GetLobbyByID(id)
				if lob.State == lobby.Waiting {
					lob.RemovePlayer(player)
				}
			})
		}
	}

}

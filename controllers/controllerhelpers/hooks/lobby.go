// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

//Package hooks contains event hooks, which are called after specific events, like lobby joins and logins.
package hooks

import (
	"fmt"
	"strconv"
	"time"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/models/chat"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

func AfterLobbyJoin(so *wsevent.Client, lob *lobby.Lobby, player *player.Player) {
	room := fmt.Sprintf("%s_private", GetLobbyRoom(lob.ID))
	//make all sockets join the private room, given the one the player joined the lobby on
	//might close, so lobbyStart and lobbyReadyUp can be sent to other tabs
	sockets, _ := sessions.GetSockets(player.SteamID)
	for _, so := range sockets {
		socket.AuthServer.Join(so, room)
	}
	if lob.State == lobby.InProgress { // player is a substitute
		lob.AfterPlayerNotInGameFunc(player, 5*time.Minute, func() {
			// if player doesn't join game server in 5 minutes,
			// substitute them
			message := player.Alias() + " has been reported for not joining the game within 5 minutes"
			chat.SendNotification(message, int(lob.ID))
			lob.Substitute(player)
		})
	}

	broadcaster.SendMessage(player.SteamID, "lobbyJoined", lobby.DecorateLobbyData(lob, false))
}

func AfterLobbyLeave(lob *lobby.Lobby, player *player.Player, kicked bool, notReady bool) {
	event := lobby.LobbyEvent{
		ID:       lob.ID,
		Kicked:   kicked,
		NotReady: notReady,
	}

	broadcaster.SendMessage(player.SteamID, "lobbyLeft", event)

	sockets, _ := sessions.GetSockets(player.SteamID)
	//player might have connected from multiple tabs, remove all of them from the room
	for _, so := range sockets {
		socket.AuthServer.Leave(so, fmt.Sprintf("%s_private", GetLobbyRoom(lob.ID)))
	}
}

func AfterLobbySpec(server *wsevent.Server, so *wsevent.Client, player *player.Player, lob *lobby.Lobby) {
	//remove socket from room of the previous lobby the socket was spectating (if any)
	lobbyID, ok := sessions.GetSpectating(so.ID)
	if ok {
		server.Leave(so, fmt.Sprintf("%d_public", lobbyID))
		sessions.RemoveSpectator(so.ID)
		if player != nil {
			prevLobby, _ := lobby.GetLobbyByID(lobbyID)
			prevLobby.RemoveSpectator(player, true)
		}
	}

	server.Join(so, fmt.Sprintf("%d_public", lob.ID))
	chelpers.BroadcastScrollback(so, lob.ID)
	sessions.SetSpectator(so.ID, lob.ID)
}

func AfterLobbySpecLeave(so *wsevent.Client, lob *lobby.Lobby) {
	socket.AuthServer.Leave(so, fmt.Sprintf("%s_public", GetLobbyRoom(lob.ID)))
	sessions.RemoveSpectator(so.ID)
}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}

//Not really broadcast, since it sends each client a different LobbyStart JSON
func BroadcastLobbyStart(lob *lobby.Lobby) {
	for _, slot := range lob.GetAllSlots() {
		player, _ := player.GetPlayerByID(slot.PlayerID)

		connectInfo := lobby.DecorateLobbyConnect(lob, player, slot.Slot)
		broadcaster.SendMessage(player.SteamID, "lobbyStart", connectInfo)
	}
}

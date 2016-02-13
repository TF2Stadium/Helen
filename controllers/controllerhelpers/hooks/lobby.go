// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

//Package hooks contains event hooks, which are called after specific events, like lobby joins and logins.
package hooks

import (
	"fmt"
	"strconv"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

func AfterLobbyJoin(so *wsevent.Client, lobby *models.Lobby, player *models.Player) {
	room := fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID))
	//make all sockets join the private room, given the one the player joined the lobby on
	//might close, so lobbyStart and lobbyReadyUp can be sent to other tabs
	sockets, _ := sessions.GetSockets(player.SteamID)
	for _, so := range sockets {
		socket.AuthServer.Join(so, room)
	}

	broadcaster.SendMessage(player.SteamID, "lobbyJoined", models.DecorateLobbyData(lobby, false))
}

func AfterLobbyLeave(lobby *models.Lobby, player *models.Player) {
	broadcaster.SendMessage(player.SteamID, "lobbyLeft", models.DecorateLobbyLeave(lobby))

	sockets, _ := sessions.GetSockets(player.SteamID)
	//player might have connected from multiple tabs, remove all of them from the room
	for _, so := range sockets {
		socket.AuthServer.Leave(so, fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID)))
	}
}

func AfterLobbySpec(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby) {
	//remove socket from room of the previous lobby the socket was spectating (if any)
	lobbyID, ok := sessions.GetSpectating(so.ID)
	if ok {
		server.Leave(so, fmt.Sprintf("%d_public", lobbyID))
		sessions.RemoveSpectator(so.ID)
	}

	server.Join(so, fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
	chelpers.BroadcastScrollback(so, lobby.ID)
	sessions.SetSpectator(so.ID, lobby.ID)
}

func AfterLobbySpecLeave(so *wsevent.Client, lobby *models.Lobby) {
	socket.AuthServer.Leave(so, fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
	sessions.RemoveSpectator(so.ID)
}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}

//Not really broadcast, since it sends each client a different LobbyStart JSON
func BroadcastLobbyStart(lobby *models.Lobby) {
	for _, slot := range lobby.GetAllSlots() {
		player, _ := models.GetPlayerByID(slot.PlayerID)

		connectInfo := models.DecorateLobbyConnect(lobby, player.Name, slot.Slot)
		broadcaster.SendMessage(player.SteamID, "lobbyStart", connectInfo)
	}
}

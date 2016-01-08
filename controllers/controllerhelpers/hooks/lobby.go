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
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

func AfterLobbyJoin(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby, player *models.Player) {
	room := fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID))
	server.AddClient(so, room)

	broadcaster.SendMessage(player.SteamID, "lobbyJoined", models.DecorateLobbyData(lobby, false))
}

func AfterLobbyLeave(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby, player *models.Player) {
	//pub := fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID))

	// bytes, _ := json.Marshal(models.DecorateLobbyData(lobby, true))
	// broadcaster.SendMessageToRoom(pub, "lobbyData", string(bytes))

	broadcaster.SendMessage(player.SteamID, "lobbyLeft", models.DecorateLobbyLeave(lobby))

	server.RemoveClient(so.Id(), fmt.Sprintf("%s_private", GetLobbyRoom(lobby.ID)))
}

func AfterLobbySpec(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby) {
	server.AddClient(so, fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
	chelpers.BroadcastScrollback(so, lobby.ID)
}

func AfterLobbySpecLeave(server *wsevent.Server, so *wsevent.Client, lobby *models.Lobby) {
	server.RemoveClient(so.Id(), fmt.Sprintf("%s_public", GetLobbyRoom(lobby.ID)))
}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}

//Not really broadcast, since it sends each client a different LobbyStart JSON
func BroadcastLobbyStart(lobby *models.Lobby) {
	var slots []*models.LobbySlot

	db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).Find(&slots)

	for _, slot := range slots {
		var player models.Player
		db.DB.First(&player, slot.PlayerID)

		_, class, _ := models.LobbyGetSlotInfoString(lobby.Type, slot.Slot)
		broadcaster.SendMessage(player.SteamID, "lobbyStart", models.DecorateLobbyConnect(lobby, player.Name, class))
	}
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/models"
	"github.com/googollee/go-socket.io"
	"strconv"
)

var BanTypeList = []string{"join", "create", "chat", "full"}

var BanTypeMap = map[string]models.PlayerBanType{
	"join":   models.PlayerBanJoin,
	"create": models.PlayerBanCreate,
	"chat":   models.PlayerBanChat,
	"full":   models.PlayerBanFull,
}

func AfterLobbyJoin(so socketio.Socket, lobby *models.Lobby, player *models.Player) {
	so.Join(GetLobbyRoom(lobby.ID))
}

func AfterLobbyLeave(so socketio.Socket, lobby *models.Lobby, player *models.Player) {
	so.Join(GetLobbyRoom(lobby.ID))

}

func AfterConnect(so socketio.Socket) {
	so.Join(config.Constants.GlobalChatRoom) //room for global chat
	models.BroadcastLobbyList()
}

func AfterConnectLoggedIn(so socketio.Socket, player *models.Player) {
	lobbyIdPlaying, err := player.GetLobbyId()
	if err == nil {
		so.Join(GetLobbyRoom(lobbyIdPlaying))
		lobby, _ := models.GetLobbyById(lobbyIdPlaying)
		models.BroadcastLobbyToUser(lobby, GetSteamId(so.Id()))
	}

	lobbyIdsSpectating, err2 := player.GetSpectatingIds()
	if err2 == nil {
		for _, id := range lobbyIdsSpectating {
			so.Join(GetLobbyRoom(id))
			lobby, _ := models.GetLobbyById(lobbyIdPlaying)
			models.BroadcastLobbyToUser(lobby, GetSteamId(so.Id()))
		}
	}
}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}

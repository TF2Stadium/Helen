// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket/internal/handler"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/googollee/go-socket.io"
)

func SocketInit(so socketio.Socket) {
	chelpers.AuthenticateSocket(so.Id(), so.Request())
	if chelpers.IsLoggedInSocket(so.Id()) {
		steamid := chelpers.GetSteamId(so.Id())
		broadcaster.SetSocket(steamid, so)
	}

	so.On("disconnection", func() {
		chelpers.DeauthenticateSocket(so.Id())
		if chelpers.IsLoggedInSocket(so.Id()) {
			steamid := chelpers.GetSteamId(so.Id())
			broadcaster.RemoveSocket(steamid)
		}
		helpers.Logger.Debug("on disconnect")
	})

	so.On("authenticationTest", chelpers.FilterRequest(so, chelpers.FilterParams{FilterLogin: true},
		func(_ map[string]interface{}) string {
			return "authenticated"
		}))

	helpers.Logger.Debug("on connection")
	chelpers.AfterConnect(so)

	loggedIn := chelpers.IsLoggedInSocket(so.Id())
	if loggedIn {
		player, err := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
		if err != nil {
			helpers.Logger.Warning("User has a cookie with but a matching player record doesn't exist: %s",
				chelpers.GetSteamId(so.Id()))
			return
		}

		chelpers.AfterConnectLoggedIn(so, player)
	} else {
		so.Emit("playerSettings", "{}")
		so.Emit("playerProfile", "{}")
	}

	// LOBBY CREATE
	so.On("lobbyCreate", handler.LobbyCreate(so))

	so.On("serverVerify", handler.ServerVerify(so))

	so.On("lobbyClose", handler.LobbyClose(so))

	so.On("lobbyJoin", handler.LobbyJoin(so))

	if loggedIn {
		so.On("lobbySpectatorJoin", handler.LobbySpectatorJoin(so))
	} else {
		so.On("lobbySpectatorJoin", handler.LobbyNoLoginSpectatorJoin(so))
	}
	so.On("lobbyKick", handler.LobbyKick(so))

	so.On("playerReady", handler.PlayerReady(so))

	so.On("playerNotReady", handler.PlayerNotReady(so))

	so.On("playerSettingsGet", handler.PlayerSettingsGet(so))

	so.On("playerSettingsSet", handler.PlayerSettingsSet(so))

	so.On("playerProfile", handler.PlayerProfile(so))

	so.On("chatSend", handler.ChatSend(so))

	so.On("adminChangeRole", handler.AdminChangeRole(so))

	so.On("requestLobbyListData", handler.RequestLobbyListData(so))

	//Debugging handlers
	if config.Constants.ServerMockUp {
		so.On("debugLobbyFill", handler.DebugLobbyFill(so))
		so.On("debugLobbyReady", handler.DebugLobbyReady(so))
		so.On("debugGetAllLobbies", handler.DebugRequestAllLobbies(so))
		so.On("debugRequestLobbyStart", handler.DebugRequestLobbyStart(so))
	}

	so.Emit("socketInitialized", "")
}

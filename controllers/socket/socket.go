package socket

import (
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/googollee/go-socket.io"
)

func SocketInit(so socketio.Socket) {
	chelpers.AuthenticateSocket(so.Id(), so.Request())
	if chelpers.IsLoggedInSocket(so.Id()) {
		steamid := chelpers.GetSteamId(so.Id())
		broadcaster.SteamIdSocketMap[steamid] = &so
	}

	so.On("disconnection", func() {
		chelpers.DeauthenticateSocket(so.Id())
		if chelpers.IsLoggedInSocket(so.Id()) {
			steamid := chelpers.GetSteamId(so.Id())
			delete(broadcaster.SteamIdSocketMap, steamid)
		}
		helpers.Logger.Debug("on disconnect")
	})

	so.On("authenticationTest", chelpers.FilterRequest(so, chelpers.FilterParams{FilterLogin: true},
		func(_ map[string]interface{}) string {
			return "authenticated"
		}))

	helpers.Logger.Debug("on connection")
	chelpers.AfterConnect(so)

	if chelpers.IsLoggedInSocket(so.Id()) {
		player, err := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
		if err != nil {
			helpers.Logger.Warning("User has a cookie with but a matching player record doesn't exist: %s",
				chelpers.GetSteamId(so.Id()))
			return
		}

		chelpers.AfterConnectLoggedIn(so, player)
	}

	// LOBBY CREATE
	so.On("lobbyCreate", lobbyCreateHandler(so))

	so.On("lobbyClose", lobbyCloseHandler(so))

	so.On("lobbyJoin", lobbyJoinHandler(so))

	so.On("lobbySpectatorJoin", lobbySpectatorJoinHandler(so))

	so.On("lobbyKick", lobbyKickHandler(so))

	so.On("playerReady", playerReadyHandler(so))

	so.On("playerUnready", playerUnreadyHandler(so))

	so.On("playerSettingsGet", playerSettingsGetHandler(so))

	so.On("playerSettingsSet", playerSettingsSetHandler(so))

	so.On("playerProfile", playerProfileHandler(so))

	so.On("chatSend", chatSendHandler(so))

	so.On("adminChangeRole", adminChangeRoleHandler(so))

	so.On("requestLobbyListData", requestLobbyListDataHandler(so))

}

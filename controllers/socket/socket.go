// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"encoding/json"
	"errors"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket/internal/handler"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

var ErrRecordNotFound = errors.New("Player record for found.")

func onDisconnect(id string) {
	//defer helpers.Logger.Debug("Disconnected from Socket")
	defer chelpers.DeauthenticateSocket(id)
	if chelpers.IsLoggedInSocket(id) {
		steamid := chelpers.GetSteamId(id)
		broadcaster.RemoveSocket(steamid)
		player, tperr := models.GetPlayerBySteamId(steamid)
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
			lobby, _ := models.GetLobbyById(id)
			err := lobby.RemoveSpectator(player, true)
			if err != nil {
				helpers.Logger.Error(err.Error())
				continue
			}
			//helpers.Logger.Debug("removing %s from %d", player.SteamId, id)
		}

	}
}

func getEvent(data []byte) string {
	var js struct {
		Request string
	}
	json.Unmarshal(data, &js)
	return js.Request
}

func ServerInit(server *wsevent.Server, noAuthServer *wsevent.Server) {
	server.OnDisconnect = onDisconnect
	server.Extractor = getEvent

	noAuthServer.OnDisconnect = onDisconnect
	noAuthServer.Extractor = getEvent

	server.On("authenticationTest", func(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
		reqerr := chelpers.FilterRequest(so, 0, true)

		if reqerr != nil {
			bytes, _ := json.Marshal(reqerr)
			return bytes
		}

		bytes, _ := json.Marshal(struct {
			Message string `json:"message"`
		}{"authenticated"})
		return bytes
	})
	//Global Handlers
	server.Register(handler.Global{})
	//Lobby Handlers
	server.Register(handler.Lobby{})
	//server.On("lobbyCreate", handler.LobbyCreate)
	//Player Handlers
	server.Register(handler.Player{})
	//Chat Handlers
	server.Register(handler.Chat{})
	//Admin Handlers
	server.Register(handler.Admin{})
	//Ban Handlers
	handler.InitializeBans(server)
	//Debugging handlers
	// if config.Constants.ServerMockUp {
	// 	server.On("debugLobbyFill", handler.DebugLobbyFill)
	// 	server.On("debugLobbyReady", handler.DebugLobbyReady)
	// 	server.On("debugUpdateStatsFilter", handler.DebugUpdateStatsFilter)
	// 	server.On("debugPlayerSub", handler.DebugPlayerSub)
	// }

	noAuthServer.On("lobbySpectatorJoin", func(s *wsevent.Server, so *wsevent.Client, data []byte) []byte {
		var args struct {
			Id *uint `json:"id"`
		}

		if err := chelpers.GetParams(data, &args); err != nil {
			return helpers.NewTPErrorFromError(err).Encode()
		}

		var lob *models.Lobby
		lob, tperr := models.GetLobbyById(*args.Id)

		if tperr != nil {
			return tperr.Encode()
		}

		chelpers.AfterLobbySpec(s, so, lob)
		bytes, _ := json.Marshal(models.DecorateLobbyData(lob, true))

		so.EmitJSON(helpers.NewRequest("lobbyData", string(bytes)))

		return chelpers.EmptySuccessJS
	})
	noAuthServer.On("getSocketInfo", (handler.Global{}).GetSocketInfo)

	noAuthServer.DefaultHandler = func(_ *wsevent.Server, so *wsevent.Client, data []byte) []byte {
		return helpers.NewTPError("Player isn't logged in.", -4).Encode()
	}
}

func SocketInit(server *wsevent.Server, noauth *wsevent.Server, so *wsevent.Client) error {
	chelpers.AuthenticateSocket(so.Id(), so.Request())
	loggedIn := chelpers.IsLoggedInSocket(so.Id())
	if loggedIn {
		steamid := chelpers.GetSteamId(so.Id())
		broadcaster.SetSocket(steamid, so)
	}

	if loggedIn {
		chelpers.AfterConnect(server, so)

		player, err := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
		if err != nil {
			helpers.Logger.Warning(
				"User has a cookie with but a matching player record doesn't exist: %s",
				chelpers.GetSteamId(so.Id()))
			chelpers.DeauthenticateSocket(so.Id())
			so.Close()
			return ErrRecordNotFound
		}

		chelpers.AfterConnectLoggedIn(server, so, player)
	} else {
		chelpers.AfterConnect(noauth, so)
		so.EmitJSON(helpers.NewRequest("playerSettings", "{}"))
		so.EmitJSON(helpers.NewRequest("playerProfile", "{}"))
	}

	so.EmitJSON(helpers.NewRequest("socketInitialized", "{}"))

	return nil
}

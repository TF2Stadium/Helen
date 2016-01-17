// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"encoding/json"
	"errors"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

var ErrRecordNotFound = errors.New("Player record for found.")

func getEvent(data []byte) string {
	var js struct {
		Request string
	}
	json.Unmarshal(data, &js)
	return js.Request
}

//SocketInit initializes the websocket connection for the provided socket
func SocketInit(so *wsevent.Client) error {
	chelpers.AuthenticateSocket(so.Id(), so.Request())
	loggedIn := chelpers.IsLoggedInSocket(so.Id())
	if loggedIn {
		steamid := chelpers.GetSteamId(so.Id())
		sessions.AddSocket(steamid, so)
	}

	if loggedIn {
		hooks.AfterConnect(AuthServer, so)

		player, err := models.GetPlayerBySteamID(chelpers.GetSteamId(so.Id()))
		if err != nil {
			helpers.Logger.Warning(
				"User has a cookie with but a matching player record doesn't exist: %s",
				chelpers.GetSteamId(so.Id()))
			chelpers.DeauthenticateSocket(so.Id())
			hooks.AfterConnect(UnauthServer, so)
			return ErrRecordNotFound
		}

		hooks.AfterConnectLoggedIn(AuthServer, so, player)
	} else {
		hooks.AfterConnect(UnauthServer, so)
		so.EmitJSON(helpers.NewRequest("playerSettings", "{}"))
		so.EmitJSON(helpers.NewRequest("playerProfile", "{}"))
	}

	so.EmitJSON(helpers.NewRequest("socketInitialized", "{}"))

	return nil
}

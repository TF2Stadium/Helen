package controllers

import (
	"errors"
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/login"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	//check if player is in the whitelist
	if config.Constants.SteamIDWhitelist != "" {
		allowed := false

		session, err := chelpers.GetSessionHTTP(r)
		if err == nil {
			steamid, ok := session.Values["steam_id"]
			allowed = ok && chelpers.IsSteamIDWhitelisted(steamid.(string))
		}
		if !allowed {
			http.Error(w, "Sorry, but you're not in the closed alpha", 403)
			return
		}
	}

	session, err := chelpers.GetSessionHTTP(r)
	var so *wsevent.Client

	if err == nil {
		_, ok := session.Values["steam_id"]
		if ok {
			so, err = socket.AuthServer.NewClient(upgrader, w, r)
		} else {
			so, err = socket.UnauthServer.NewClient(upgrader, w, r)
		}
	} else {
		var estr = "Couldn't create WebSocket connection."
		//estr = err.Error()

		helpers.Logger.Error(err.Error())
		http.Error(w, estr, 500)
		return
	}

	if err != nil || so == nil {
		login.LogoutSession(w, r)
		return
	}

	//helpers.Logger.Debug("Connected to Socket")
	err = SocketInit(so)
	if err != nil {
		login.LogoutSession(w, r)
	}
}

var ErrRecordNotFound = errors.New("Player record for found.")

//SocketInit initializes the websocket connection for the provided socket
func SocketInit(so *wsevent.Client) error {
	chelpers.AuthenticateSocket(so.Id(), so.Request())
	loggedIn := chelpers.IsLoggedInSocket(so.Id())
	if loggedIn {
		steamid := chelpers.GetSteamId(so.Id())
		sessions.AddSocket(steamid, so)
	}

	if loggedIn {
		hooks.AfterConnect(socket.AuthServer, so)

		player, err := models.GetPlayerBySteamID(chelpers.GetSteamId(so.Id()))
		if err != nil {
			helpers.Logger.Warning(
				"User has a cookie with but a matching player record doesn't exist: %s",
				chelpers.GetSteamId(so.Id()))
			chelpers.DeauthenticateSocket(so.Id())
			hooks.AfterConnect(socket.UnauthServer, so)
			return ErrRecordNotFound
		}

		hooks.AfterConnectLoggedIn(so, player)
	} else {
		hooks.AfterConnect(socket.UnauthServer, so)
		so.EmitJSON(helpers.NewRequest("playerSettings", "{}"))
		so.EmitJSON(helpers.NewRequest("playerProfile", "{}"))
	}

	so.EmitJSON(helpers.NewRequest("socketInitialized", "{}"))

	return nil
}

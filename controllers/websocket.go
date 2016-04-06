package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	token, err := chelpers.GetToken(r)
	if err != nil && err != http.ErrNoCookie { //invalid jwt token
		token = nil
	}

	//check if player is in the whitelist
	if config.Constants.SteamIDWhitelist != "" {
		if token == nil {
			// player isn't logged in,
			// and access is restricted to logged in people
			http.Error(w, "Not logged in", http.StatusForbidden)
			return
		}

		if !chelpers.IsSteamIDWhitelisted(token.Claims["steam_id"].(string)) {
			http.Error(w, "you're not in the beta", http.StatusForbidden)
			return
		}
	}

	var so *wsevent.Client

	if token != nil { //received valid jwt
		so, err = socket.AuthServer.NewClient(upgrader, w, r)
	} else {
		so, err = socket.UnauthServer.NewClient(upgrader, w, r)
	}

	if err != nil {
		return
	}

	so.Token = token

	//logrus.Debug("Connected to Socket")
	err = SocketInit(so)
	if err != nil {
		logrus.Error(err)
		so.Close()
		return
	}
}

var ErrRecordNotFound = errors.New("Player record for found.")

//SocketInit initializes the websocket connection for the provided socket
func SocketInit(so *wsevent.Client) error {
	loggedIn := so.Token != nil

	if loggedIn {
		hooks.AfterConnect(socket.AuthServer, so)

		steamid := so.Token.Claims["steam_id"].(string)

		player, err := player.GetPlayerBySteamID(steamid)
		if err != nil {
			return fmt.Errorf("Couldn't find player record for %s", steamid)
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

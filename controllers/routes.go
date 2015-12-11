// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket"
	"github.com/TF2Stadium/wsevent"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

type Sockets struct {
	Auth   *wsevent.Server
	Noauth *wsevent.Server
}

func (s Sockets) SocketHandler(w http.ResponseWriter, r *http.Request) {
	if config.Constants.SteamIDWhitelist != "" {
		session, err := chelpers.GetSessionHTTP(r)

		allowed := true

		if err == nil {
			steamid, ok := session.Values["steam_id"]
			if !ok {
				allowed = false
			} else if !chelpers.IsSteamIDWhitelisted(steamid.(string)) {
				allowed = false
			}
		} else {
			allowed = false
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
			so, err = s.Auth.NewClient(upgrader, w, r)
		} else {
			so, err = s.Noauth.NewClient(upgrader, w, r)
		}
	} else {
		var estr = "Couldn't create WebSocket connection."
		//estr = err.Error()

		http.Error(w, estr, 500)
		return
	}

	if err != nil || so == nil {
		LogoutSession(w, r)
		return
	}

	//helpers.Logger.Debug("Connected to Socket")
	err = socket.SocketInit(s.Auth, s.Noauth, so)
	if err != nil {
		LogoutSession(w, r)
	}
}

func SetupHTTPRoutes(server *wsevent.Server, noauth *wsevent.Server) {
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/openidcallback", LoginCallbackHandler)
	http.HandleFunc("/startLogin", LoginHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/chatlogs/", GetChatLogs)
	if config.Constants.MockupAuth {
		http.HandleFunc("/startMockLogin/", MockLoginHandler)
	}
	http.HandleFunc("/websocket/", Sockets{server, noauth}.SocketHandler)
}

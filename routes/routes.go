// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package routes

import (
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket"
	"github.com/TF2Stadium/wsevent"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

func SetupHTTPRoutes(server *wsevent.Server, noauth *wsevent.Server) {
	http.HandleFunc("/", controllers.MainHandler)
	http.HandleFunc("/openidcallback", controllers.LoginCallbackHandler)
	http.HandleFunc("/startLogin", controllers.LoginHandler)
	http.HandleFunc("/logout", controllers.LogoutHandler)
	if config.Constants.MockupAuth {
		http.HandleFunc("/startMockLogin/", controllers.MockLoginHandler)
	}
	http.HandleFunc("/websocket/", func(w http.ResponseWriter, r *http.Request) {
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
			steamid, ok := session.Values["steam_id"]
			if ok {
				so, err = server.NewClientWithID(upgrader, w, r, steamid.(string))
			} else {
				so, err = noauth.NewClient(upgrader, w, r)
			}
		} else {
			var estr = "Couldn't create WebSocket connection."
			if err != nil {
				estr = err.Error()
			}

			http.Error(w, estr, 500)
			return
		}

		//helpers.Logger.Debug("Connected to Socket")
		err = socket.SocketInit(server, noauth, so)
		if err != nil {
			controllers.LogoutHandler(w, r)
		}
	})
}

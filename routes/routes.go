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

func SetupHTTPRoutes(server *wsevent.Server) {
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
				if _, ok := session.Values["steam_id"]; !ok {
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

		so, err := server.NewClient(upgrader, w, r)

		//helpers.Logger.Debug("Connected to Socket")
		err = socket.SocketInit(server, so)
		if err != nil {
			controllers.LogoutHandler(w, r)
		}
	})
}

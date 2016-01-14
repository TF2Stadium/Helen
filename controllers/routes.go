// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/admin"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/wsevent"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}

type Sockets struct {
	Auth   *wsevent.Server
	Noauth *wsevent.Server
}

func (s Sockets) SocketHandler(w http.ResponseWriter, r *http.Request) {
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
			so, err = s.Auth.NewClient(upgrader, w, r)
		} else {
			so, err = s.Noauth.NewClient(upgrader, w, r)
		}
	} else {
		var estr = "Couldn't create WebSocket connection."
		//estr = err.Error()

		helpers.Logger.Error(err.Error())
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

type route struct {
	pattern string
	handler func(http.ResponseWriter, *http.Request)
}

func SetupHTTPRoutes(mux *http.ServeMux, server *wsevent.Server, noauth *wsevent.Server) {

	var routes = []route{
		{"/", MainHandler},
		{"/openidcallback", LoginCallbackHandler},
		{"/startLogin", LoginHandler},
		{"/logout", LogoutHandler},
		{"/websocket/", Sockets{server, noauth}.SocketHandler},

		{"/admin", chelpers.FilterHTTPRequest(helpers.ActionViewPage, admin.ServeAdminPage)},

		{"/admin/roles", chelpers.FilterHTTPRequest(helpers.ActionViewPage, admin.ServeAdminRolePage)},
		{"/admin/roles/addadmin", chelpers.FilterHTTPRequest(helpers.ActionChangeRole, admin.AddAdmin)},
		{"/admin/roles/addmod", chelpers.FilterHTTPRequest(helpers.ActionChangeRole, admin.AddMod)},
		{"/admin/roles/remove", chelpers.FilterHTTPRequest(helpers.ActionChangeRole, admin.Remove)},
		{"/admin/roles/adddev", chelpers.FilterHTTPRequest(helpers.ActionChangeRole, admin.AddDeveloper)},

		{"/admin/ban", chelpers.FilterHTTPRequest(helpers.ActionViewPage, admin.ServeAdminBanPage)},
		{"/admin/ban/join", chelpers.FilterHTTPRequest(helpers.ActionBanJoin, admin.BanJoin)},
		{"/admin/ban/create", chelpers.FilterHTTPRequest(helpers.ActionBanCreate, admin.BanCreate)},
		{"/admin/ban/chat", chelpers.FilterHTTPRequest(helpers.ActionBanChat, admin.BanChat)},

		{"/admin/chatlogs", chelpers.FilterHTTPRequest(helpers.ActionViewLogs, admin.GetChatLogs)},
	}

	if config.Constants.MockupAuth {
		mux.HandleFunc("/startMockLogin/", MockLoginHandler)
	}

	for _, route := range routes {
		mux.HandleFunc(route.pattern, route.handler)
	}
}

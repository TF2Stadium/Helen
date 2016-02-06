// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package routes

import (
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers"
	"github.com/TF2Stadium/Helen/controllers/admin"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/login"
	"github.com/TF2Stadium/Helen/helpers"
)

type route struct {
	pattern string
	handler func(http.ResponseWriter, *http.Request)
}

var routes = []route{
	{"/", controllers.MainHandler},
	{"/openidcallback", login.LoginCallbackHandler},
	{"/startLogin", login.LoginHandler},
	{"/startTwitchLogin", login.TwitchLogin},
	{"/twitchAuth", login.TwitchAuth},
	{"/logout", login.LogoutHandler},
	{"/websocket/", controllers.SocketHandler},

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
	{"/admin/banlogs", chelpers.FilterHTTPRequest(helpers.ActionViewLogs, admin.DisplayLogs)},
}

func SetupHTTP(mux *http.ServeMux) {
	if config.Constants.MockupAuth {
		mux.HandleFunc("/startMockLogin/", login.MockLoginHandler)
	}

	for _, route := range routes {
		mux.HandleFunc(route.pattern, route.handler)
	}
}

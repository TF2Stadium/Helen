// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"html/template"
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	"golang.org/x/net/xsrftoken"
)

var banForm = map[string]string{
	"joinLobby":   "Joining Lobbies",
	"createLobby": "Creating Lobbies",
	"chat":        "Chatting",
	"full":        "Full ban",
}

var roleForm = map[string]string{
	"admin": "Add Administrator",
	"mod":   "Add Moderator",
}

var adminPagetempl *template.Template

func init() {
	adminPagetempl, _ = template.ParseFiles("./views/admin/index.html")
}

func ServeAdminPage(w http.ResponseWriter, r *http.Request) {
	adminPagetempl.Execute(w, map[string]interface{}{
		"BanForms":  banForm,
		"RoleForms": roleForm,
		"XSRFToken": xsrftoken.Generate(config.Constants.CookieStoreSecret, "admin", "POST"),
		"SiteKey":   config.Constants.ReCaptchaSiteKey,
	})
}

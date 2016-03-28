// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"golang.org/x/net/xsrftoken"
)

var banForm = map[string]string{
	"joinLobby":       "Joining Lobbies",
	"joinMumbleLobby": "Joining Mumble Lobbies",
	"createLobby":     "Creating Lobbies",
	"chat":            "Chatting",
	"full":            "Full ban",
}

var roleForm = map[string]string{
	"admin": "Add Administrator",
	"mod":   "Add Moderator",
}

var adminPageTempl *template.Template

func ServeAdminPage(w http.ResponseWriter, r *http.Request) {
	err := adminPageTempl.Execute(w, map[string]interface{}{
		"BanForms":  banForm,
		"RoleForms": roleForm,
		"XSRFToken": xsrftoken.Generate(config.Constants.CookieStoreSecret, "admin", "POST"),
	})
	if err != nil {
		logrus.Error(err)
	}
}

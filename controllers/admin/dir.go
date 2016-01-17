// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/TF2Stadium/Helen/helpers"
)

func ServeAdminPage(w http.ResponseWriter, r *http.Request) {
	abs, _ := filepath.Abs("./views/admin")
	http.ServeFile(w, r, abs)
}

type ban struct {
	URL   string //URL
	Title string //Title
}

var bans = []ban{
	{"ban/join", "Ban from joining lobbies"},
	{"ban/create", "Ban from creating lobbies"},
	{"ban/chat", "Ban from chatting"},
	{"ban/full", "Full ban"},
}

func ServeAdminBanPage(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("views/admin/templates/ban_forms.html")
	if err != nil {
		helpers.Logger.Error(err.Error())
		return
	}

	templ.Execute(w, bans)
}

func ServeAdminRolePage(w http.ResponseWriter, r *http.Request) {
	abs, _ := filepath.Abs("./views/admin/roles")
	http.ServeFile(w, r, abs)
}

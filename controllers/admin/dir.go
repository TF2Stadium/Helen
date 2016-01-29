// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/Sirupsen/logrus"
)

type form struct {
	URL   string //URL
	Title string //Title
}

var banForm = []form{
	{"ban/join", "Ban from joining lobbies"},
	{"ban/create", "Ban from creating lobbies"},
	{"ban/chat", "Ban from chatting"},
	{"ban/full", "Full ban"},
}

var roleForm = []form{
	{"roles/addadmin", "Add Admin"},
	{"roles/addmod", "Add Mod"},
	{"roles/adddev", "Add Developer"},

	{"roles/remove", "Remove Admin/Mod"},
}

func ServeAdminPage(w http.ResponseWriter, r *http.Request) {
	abs, _ := filepath.Abs("./views/admin")
	http.ServeFile(w, r, abs)
}

func ServeAdminBanPage(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("views/admin/templates/ban_forms.html")
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	templ.Execute(w, banForm)
}

func ServeAdminRolePage(w http.ResponseWriter, r *http.Request) {
	templ, err := template.ParseFiles("views/admin/templates/role_forms.html")
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	templ.Execute(w, roleForm)
}

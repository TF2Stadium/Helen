package admin

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/models/gameserver"
	"golang.org/x/net/xsrftoken"
)

var serverPage *template.Template

func AddServer(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	values := r.Form

	token := values.Get("xsrf-token")
	if !xsrftoken.Valid(token, config.Constants.CookieStoreSecret, "admin", "POST") {
		http.Error(w, "invalid xsrf token", http.StatusBadRequest)
		return
	}

	name := values.Get("name")
	if name == "" {
		http.Error(w, "Empty name not allowed", http.StatusBadRequest)
		return
	}

	addr := values.Get("address")
	if addr == "" {
		http.Error(w, "Empty address not allowed", http.StatusBadRequest)
		return
	}

	passwd := values.Get("password")
	if passwd == "" {
		http.Error(w, "Empty password not allowed", http.StatusBadRequest)
		return
	}

	server, err := gameserver.NewStoredServer(name, addr, passwd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Server successfully added (ID: #%d)", server.ID)
}

func RemoveServer(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	values := r.Form

	token := values.Get("xsrf-token")
	if !xsrftoken.Valid(token, config.Constants.CookieStoreSecret, "admin", "POST") {
		http.Error(w, "invalid xsrf token", http.StatusBadRequest)
		return
	}

	addr := values.Get("address")
	if addr == "" {
		http.Error(w, "Empty address not allowed", http.StatusBadRequest)
		return
	}

	gameserver.RemoveStoredServer(addr)
	fmt.Fprintf(w, "Server successfully deleted.")
}

func ViewServerPage(w http.ResponseWriter, r *http.Request) {
	err := serverPage.Execute(w, map[string]interface{}{
		"XSRFToken": xsrftoken.Generate(config.Constants.CookieStoreSecret, "admin", "POST"),
		"Servers":   gameserver.GetAllStoredServers(),
	})
	if err != nil {
		logrus.Error(err)
	}
}

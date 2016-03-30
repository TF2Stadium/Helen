// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"fmt"
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models/player"
	"golang.org/x/net/xsrftoken"
)

func ChangeRole(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	values := r.Form
	steamid := values.Get("steamid")
	remove := values.Get("remove")
	token := values.Get("xsrf-token")
	if !xsrftoken.Valid(token, config.Constants.CookieStoreSecret, "admin", "POST") {
		http.Error(w, "invalid xsrf token", http.StatusBadRequest)
		return
	}

	role, ok := map[string]authority.AuthRole{
		"admin": helpers.RoleAdmin,
		"mod":   helpers.RoleMod,
	}[values.Get("role")]
	if !ok {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	player, err := player.GetPlayerBySteamID(steamid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if remove == "true" {
		player.Role = 0
		player.Save()
		fmt.Fprintf(w, "Player %s (%s) has been removed as %s", player.Name, player.SteamID, helpers.RoleNames[role])
		return
	}

	player.Role = role
	player.Save()
	fmt.Fprintf(w, "Player %s (%s) has been made a %s", player.Name, player.SteamID, helpers.RoleNames[role])
	return
}

func Remove(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	steamid := r.Form.Get("steamid")
	player, err := player.GetPlayerBySteamID(steamid)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	player.Role = authority.AuthRole(0)
	player.Save()
	fmt.Fprintf(w, "%s (%s) is no longer an admin/mod", player.Name, player.SteamID)
}

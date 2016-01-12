// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"fmt"
	"net/http"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
)

func AddAdmin(w http.ResponseWriter, r *http.Request) {
	steamid := r.URL.Query().Get("steamid")
	player, err := models.GetPlayerBySteamID(steamid)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	player.Role = helpers.RoleAdmin
	player.Save()
	fmt.Fprintf(w, "%s (%s) has been made an admin", player.Name, player.SteamID)
}

func AddMod(w http.ResponseWriter, r *http.Request) {
	steamid := r.URL.Query().Get("steamid")
	player, err := models.GetPlayerBySteamID(steamid)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	player.Role = helpers.RoleMod
	player.Save()
	fmt.Fprintf(w, "%s (%s) has been made a mod", player.Name, player.SteamID)
}

func Remove(w http.ResponseWriter, r *http.Request) {
	steamid := r.URL.Query().Get("steamid")
	player, err := models.GetPlayerBySteamID(steamid)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	player.Role = authority.AuthRole(0)
	player.Save()
	fmt.Fprintf(w, "%s (%s) is no longer an admin/mod", player.Name, player.SteamID)
}

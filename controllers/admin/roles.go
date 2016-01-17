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
	"github.com/TF2Stadium/Helen/models"
)

func changeRole(w http.ResponseWriter, r *http.Request, role authority.AuthRole) error {
	player, err := models.GetPlayerBySteamID(r.URL.Query().Get("steamid"))
	if err != nil {
		return err
	}

	switch r.URL.Query().Get("confirm") {
	case "yes":
		player.Role = role
		player.Save()
		fmt.Fprintf(w, "%s (%s) has been made a %s", player.Name, player.SteamID, helpers.RoleNames[role])
	default:
		query := r.URL.Query()
		query.Set("confirm", "yes")
		r.URL.RawQuery = query.Encode()
		title := fmt.Sprintf("Make %s (%s) a %s?", player.Name, player.SteamID, helpers.RoleNames[role])
		confirmReq(w, r.URL.String(), config.Constants.Domain+"/admin/roles", title)
	}

	return nil
}

func AddAdmin(w http.ResponseWriter, r *http.Request) {
	err := changeRole(w, r, helpers.RoleAdmin)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
}

func AddMod(w http.ResponseWriter, r *http.Request) {
	err := changeRole(w, r, helpers.RoleMod)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
}

func AddDeveloper(w http.ResponseWriter, r *http.Request) {
	err := changeRole(w, r, helpers.RoleDeveloper)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
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

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/models/player"
	"golang.org/x/net/xsrftoken"
)

var banlogsTempl *template.Template

func BanPlayer(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	values := r.Form
	steamid := values.Get("steamid")
	reason := values.Get("reason")
	banType := values.Get("type")
	remove := values.Get("remove")
	token := values.Get("xsrf-token")
	if !xsrftoken.Valid(token, config.Constants.CookieStoreSecret, "admin", "POST") {
		http.Error(w, "invalid xsrf token", http.StatusBadRequest)
		return
	}

	ban, ok := map[string]player.BanType{
		"joinLobby":       player.BanJoin,
		"joinMumbleLobby": player.BanJoinMumble,
		"createLobby":     player.BanCreate,
		"chat":            player.BanChat,
		"full":            player.BanFull,
	}[banType]
	if !ok {
		http.Error(w, "Invalid ban type", http.StatusBadRequest)
		return
	}

	player, err := player.GetPlayerBySteamID(steamid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if remove == "true" {
		err := player.Unban(ban)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			fmt.Fprintf(w, "Player %s (%s) has been unbanned (%s)", player.Name, player.SteamID, ban.String())
		}
		return
	}

	until, err := time.Parse("2006-01-02 15:04", values.Get("date")+" "+values.Get("time"))
	if err != nil {
		http.Error(w, "invalid time format", http.StatusBadRequest)
		return
	} else if until.Sub(time.Now()) < 0 {
		http.Error(w, "invalid time", http.StatusBadRequest)
		return
	}

	jwt, _ := chelpers.GetToken(r)
	bannedByPlayer := chelpers.GetPlayer(jwt)

	err = player.BanUntil(until, ban, reason, bannedByPlayer.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Player %s (%s) has been banned (%s) till %v", player.Name, player.SteamID, ban.String(), until)
}

func GetBanLogs(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	if !xsrftoken.Valid(values.Get("xsrf-token"), config.Constants.CookieStoreSecret, "admin", "POST") {
		http.Error(w, "invalid xsrf token", http.StatusBadRequest)
		return
	}

	var bans []*player.PlayerBan

	all := values.Get("all")

	steamid := values.Get("steamid")
	if steamid == "" {
		if all == "" {
			bans = player.GetAllActiveBans()
		} else {
			bans = player.GetAllBans()
		}

	} else {
		player, err := player.GetPlayerBySteamID(steamid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if all == "" {
			bans, _ = player.GetActiveBans()
		} else {
			bans, _ = player.GetAllBans()
		}

	}

	err := banlogsTempl.Execute(w, bans)
	if err != nil {
		logrus.Error(err)
	}
}

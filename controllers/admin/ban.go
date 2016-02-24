// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/TF2Stadium/Helen/models"
)

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
	ban, ok := map[string]models.PlayerBanType{
		"joinLobby":   models.PlayerBanJoin,
		"createLobby": models.PlayerBanCreate,
		"chat":        models.PlayerBanChat,
		"full":        models.PlayerBanFull,
	}[banType]
	if !ok {
		http.Error(w, "Invalid ban type", http.StatusBadRequest)
		return
	}

	player, err := models.GetPlayerBySteamID(steamid)
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

	err = player.BanUntil(until, ban, reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Player %s (%s) has been banned (%s) till %v", player.Name, player.SteamID, ban.String(), until)
}

var banlogsTempl = template.Must(template.ParseFiles("views/admin/templates/ban_logs.html"))

type BanData struct {
	Player *models.Player
	Ban    *models.PlayerBan
}

func GetBanLogs(w http.ResponseWriter, r *http.Request) {
	allBans := models.GetAllActiveBans()
	var bans []BanData

	for _, ban := range allBans {
		player, _ := models.GetPlayerByID(ban.PlayerID)
		banData := BanData{
			Player: player,
			Ban:    ban}

		bans = append(bans, banData)
	}

	banlogsTempl.Execute(w, bans)
}

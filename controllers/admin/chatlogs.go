// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
	"github.com/jinzhu/gorm"
)

var (
	dateRegex     = regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})`)
	chatLogsTempl *template.Template
)

func init() {
	chatLogsTempl, _ = template.ParseFiles("views/admin/templates/chatlogs.html")
}

func getPlayerID(steamID string) (playerID uint) {
	db.DB.Table("players").Select("id").Where("steam_id = ?", steamID).Row().Scan(&playerID)
	return
}

func GetChatLogs(w http.ResponseWriter, r *http.Request) {
	var messages []*models.ChatMessage
	values := r.URL.Query()

	room, err := strconv.Atoi(values.Get("room"))
	if err != nil && values.Get("room") != "" {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	steamID := values.Get("steamid")
	var from, to time.Time

	if values.Get("from") != "" { //2006-01-02
		from, err = time.Parse("2006-01-02", values.Get("from"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		from = time.Time{}
	}

	if values.Get("to") != "" {
		to, err = time.Parse("2006-01-02", values.Get("to"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		to = time.Now()
	}

	order := values.Get("order")
	var results *gorm.DB

	if values.Get("room") == "" { //Retrieve all messages sent by a specific player
		if steamID == "" {
			http.Error(w, "No Steam ID given.", http.StatusBadRequest)
			return
		}

		playerID := getPlayerID(steamID)
		if playerID == 0 {
			http.Error(w, fmt.Sprintf("Couldn't find player with Steam ID %s", steamID), http.StatusNotFound)
			return
		}

		results = db.DB.Where("player_id = ? AND room = ? AND created_at >= ? AND created_at <= ?", playerID, room, from, to)
	} else if steamID == "" { //Retrieve all messages sent to a specfic room
		results = db.DB.Where("room = ? AND (created_at >= ? AND created_at <= ?)", room, from, to)
	} else { //Retrieve all messages sent to a specific room and a speficic player
		playerID := getPlayerID(steamID)
		if playerID == 0 {
			http.Error(w, fmt.Sprintf("Couldn't find player with Steam ID %s", steamID), http.StatusNotFound)
			return
		}

		results = db.DB.Where("player_id = ? AND room = ? AND created_at >= ? AND created_at <= ?", playerID, room, from, to)
	}

	if order == "Ascending" {
		err = results.Order("id").Find(&messages).Error
	} else if order == "Descending" {
		err = results.Order("id desc").Find(&messages).Error
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, message := range messages {
		//err := db.DB.DB().QueryRow("SELECT name, profileurl FROM players WHERE id = $1", message.PlayerID).Scan(&message.Player.Name, &message.Player.ProfileURL)
		db.DB.DB().QueryRow("SELECT name, profileurl FROM players WHERE id = $1", message.PlayerID).Scan(&message.Player.Name, &message.Player.Profileurl)
	}

	err = chatLogsTempl.Execute(w, messages)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
}

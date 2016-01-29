// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"errors"
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
	dateRegex = regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})`)
)

func getPlayerID(steamID string) (playerID uint) {
	db.DB.Table("players").Select("id").Where("steam_id = ?", steamID).Row().Scan(&playerID)
	return
}

func GetChatLogs(w http.ResponseWriter, r *http.Request) {
	chatLogsTempl, err := template.ParseFiles("views/admin/templates/chatlogs.html")
	if err != nil {
		logrus.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var messages []*models.ChatMessage
	values := r.URL.Query()

	room, err := strconv.Atoi(values.Get("room"))
	if err != nil && values.Get("room") != "" {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	steamID := values.Get("steamid")
	var from, to time.Time

	if values.Get("from") != "" {
		from, err = timestamp(values.Get("from"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		from = time.Time{}
	}

	if values.Get("to") != "" {
		to, err = timestamp(values.Get("to"))
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
		err := db.DB.DB().QueryRow("SELECT name, profileurl FROM players WHERE id = $1", message.PlayerID).Scan(&message.Player.Name, &message.Player.ProfileURL)
		if err != nil {
			logrus.Warning(err.Error())
		}
	}

	err = chatLogsTempl.Execute(w, messages)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
}

//date regex - MM-DD-YYYY
func timestamp(date string) (time.Time, error) {
	if !dateRegex.MatchString(date) {
		return time.Time{}, errors.New("timestamp: invalid date")
	}

	var t time.Time
	matches := dateRegex.FindStringSubmatch(date)

	month, err := strconv.Atoi(matches[1])
	if err != nil {
		return t, err
	}

	day, err := strconv.Atoi(matches[2])
	if err != nil {
		return t, err
	}

	year, err := strconv.Atoi(matches[3])
	if err != nil {
		return t, err
	}

	t = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	return t, nil
}

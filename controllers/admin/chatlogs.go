// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"errors"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

var (
	dateRegex = regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})`)
)

func GetChatLogs(w http.ResponseWriter, r *http.Request) {
	chatLogsTempl, err := template.ParseFiles("views/admin/templates/chatlogs.html")
	if err != nil {
		helpers.Logger.Error(err.Error())
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

	table := db.DB.Table("chat_messages")
	if values.Get("room") == "" {
		if steamID == "" {
			http.Error(w, "No Steam ID given.", http.StatusBadRequest)
			return
		}
		table.Joins("INNER JOIN players ON players.id = chat_messages.player_id").Where("players.steam_id = ? AND chat_messages.created_at >= ? AND chat_messages.created_at <= ?", steamID, from, to).Find(&messages)
	} else if steamID == "" {
		table.Where("room = ? AND created_at >= ? AND created_at <= ?", room, from, to).Find(&messages)
	}

	for _, message := range messages {
		err := db.DB.DB().QueryRow("SELECT name, profileurl FROM players WHERE id = $1", message.PlayerID).Scan(&message.Player.Name, &message.Player.ProfileURL)
		if err != nil {
			helpers.Logger.Warning(err.Error())
		}
	}

	err = chatLogsTempl.Execute(w, messages)
	if err != nil {
		helpers.Logger.Error(err.Error())
		return
	}
}

func timestamp(date string) (time.Time, error) {
	if !dateRegex.MatchString(date) {
		return time.Time{}, errors.New("timestamp: invalid date")
	}

	matches := dateRegex.FindStringSubmatch(date)

	month, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, err
	}

	day, err := strconv.Atoi(matches[2])
	if err != nil {
		return time.Time{}, err
	}

	year, err := strconv.Atoi(matches[3])
	if err != nil {
		return time.Time{}, err
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	return t, nil
}

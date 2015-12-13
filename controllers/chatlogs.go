// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
)

var dateRegex = regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})`)

func GetChatLogs(w http.ResponseWriter, r *http.Request) {
	var messages []*models.ChatMessage
	steamid := strings.Index(r.URL.Path, "steamid/")
	room := strings.Index(r.URL.Path, "room/")

	if steamid != -1 {
		var err error

		steamid := r.URL.Path[strings.Index(r.URL.Path, "steamid/")+8:]
		index := strings.Index(steamid, "/")
		if index != -1 {
			steamid = steamid[:index]
		}

		player, tperr := models.GetPlayerBySteamId(steamid)
		if tperr != nil {
			http.Error(w, tperr.Error(), 400)
			return
		}

		messages, err = models.GetPlayerMessages(player)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

	} else if room != -1 {
		roomstr := r.URL.Path[strings.Index(r.URL.Path, "room/")+5:]
		index := strings.Index(roomstr, "/")
		if index != -1 {
			roomstr = roomstr[:index]
		}

		room, err := strconv.Atoi(roomstr)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		messages, err = models.GetRoomMessages(room)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}

	from := int64(0)
	to := time.Now().Unix()

	if dateRegex.MatchString(r.URL.Query().Get("from")) {
		t, err := timestamp(r.URL.Query().Get("from"))

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		from = t.Unix()
	}
	if dateRegex.MatchString(r.URL.Query().Get("to")) {
		t, err := timestamp(r.URL.Query().Get("to"))

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		to = t.Unix()
	}

	//fmt.Printf("%d %d\n", from, to)
	var filteredMessages []*models.ChatMessage
	for _, message := range messages {
		if message.CreatedAt.Unix() >= from && message.CreatedAt.Unix() <= to {
			filteredMessages = append(filteredMessages, message)
		}
	}

	logs := "<body>\n"
	format := "<font color=\"red\">[%s]</font> <a href=\"https://steamcommunity.com/profiles/%s\">%s</a>: %s<br>\n"
	if steamid != -1 {
		format = "<font color=\"red\">[%s]</font> <a href=\"https://steamcommunity.com/profiles/%s\">%s</a>: %s<br>\n"
	}

	prevRoom := -1
	//format := "%s: %s\t[%s]\n"
	for _, message := range filteredMessages {
		var player models.Player
		if prevRoom != message.Room {
			logs += fmt.Sprintf("<font color=\"blue\"> Room #%d </font><br>\n", message.Room)
		}

		prevRoom = message.Room
		db.DB.First(&player, message.PlayerID)

		if steamid != -1 {
			logs += fmt.Sprintf(format, message.CreatedAt.Format(time.RFC822), player.SteamId, player.Name, html.EscapeString(message.Message))
			continue
		}

		logs += fmt.Sprintf(format, message.CreatedAt.Format(time.RFC822), player.SteamId, player.Name, html.EscapeString(message.Message))
	}

	logs += "</body>"
	fmt.Fprint(w, logs)
}

func timestamp(date string) (*time.Time, error) {
	matches := dateRegex.FindStringSubmatch(date)

	month, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}

	day, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}

	year, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, err
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	return &t, nil
}

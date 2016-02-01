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
	"github.com/TF2Stadium/Helen/models"
)

var timeRe = regexp.MustCompile(`(\d+[a-z])`)
var banString = map[models.PlayerBanType]string{
	models.PlayerBanJoin:   "joining lobbies",
	models.PlayerBanCreate: "creating lobbies",
	models.PlayerBanChat:   "chatting",
	models.PlayerBanFull:   "the website",
}

//(y)ear (m)onth (w)eek (d)ay (h)our
func parseTime(str string) (*time.Time, error) {
	var year, month, week, day, hour int

	if !timeRe.MatchString(str) {
		return nil, errors.New("Invalid time duration")
	}

	matches := timeRe.FindStringSubmatch(str)
	for _, match := range matches {
		suffix := match[len(match)-1]
		prefix := match[:len(match)-1]
		num, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, err
		}

		switch suffix {
		case 'y':
			year = num
		case 'm':
			month = num
		case 'w':
			week = num
		case 'd':
			day = num
		case 'h':
			hour = num
		}
	}

	t := time.Now().AddDate(year, month, week*7+day).Add(time.Hour * time.Duration(hour))
	return &t, nil
}

func banPlayer(w http.ResponseWriter, r *http.Request, banType models.PlayerBanType) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	values := r.Form
	confirm := values.Get("confirm")
	steamid := values.Get("steamid")
	reason := values.Get("reason")

	player, tperr := models.GetPlayerBySteamID(steamid)
	if tperr != nil {
		return tperr
	}

	switch confirm {
	case "yes":
		if err := verifyToken(r, "banPlayer"); err != nil {
			return err
		}

		until, err := parseTime(values.Get("until"))
		if err != nil {
			return err
		}

		player.BanUntil(*until, banType, reason)
	default:
		title := fmt.Sprintf("Ban %s (%s) from %s?", player.Name, player.SteamID, banString[banType])
		confirmReq(w, r, "banPlayer", title)
	}

	return nil
}

func BanJoin(w http.ResponseWriter, r *http.Request) {
	err := banPlayer(w, r, models.PlayerBanJoin)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func BanChat(w http.ResponseWriter, r *http.Request) {
	err := banPlayer(w, r, models.PlayerBanChat)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func BanCreate(w http.ResponseWriter, r *http.Request) {
	err := banPlayer(w, r, models.PlayerBanCreate)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func BanFull(w http.ResponseWriter, r *http.Request) {
	err := banPlayer(w, r, models.PlayerBanFull)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

type BanData struct {
	Player  *models.Player
	BanName string
	Ban     *models.PlayerBan
}

func DisplayLogs(w http.ResponseWriter, r *http.Request) {
	allBans := models.GetAllActiveBans()
	var bans []BanData

	templ, err := template.ParseFiles("views/admin/templates/ban_logs.html")
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	for _, ban := range allBans {
		player, _ := models.GetPlayerByID(ban.PlayerID)
		banData := BanData{
			Player:  player,
			BanName: fmt.Sprintf("Banned from %s", banString[ban.Type]),
			Ban:     ban}

		bans = append(bans, banData)
	}

	templ.Execute(w, bans)
}

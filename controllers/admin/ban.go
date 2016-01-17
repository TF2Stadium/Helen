// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/TF2Stadium/Helen/models"
)

var timeRe = regexp.MustCompile(`(\d+[a-z])`)

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

func banPlayer(u *url.URL, banType models.PlayerBanType) error {
	values := u.Query()
	steamid := values.Get("steamid")
	reason := values.Get("reason")

	player, tperr := models.GetPlayerBySteamID(steamid)
	if tperr != nil {
		return tperr
	}

	until, err := parseTime(values.Get("until"))
	if err != nil {
		return err
	}

	player.BanUntil(*until, models.PlayerBanJoin, reason)
	return nil
}

func BanJoin(w http.ResponseWriter, r *http.Request) {
	err := banPlayer(r.URL, models.PlayerBanJoin)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func BanChat(w http.ResponseWriter, r *http.Request) {
	err := banPlayer(r.URL, models.PlayerBanChat)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func BanCreate(w http.ResponseWriter, r *http.Request) {
	err := banPlayer(r.URL, models.PlayerBanCreate)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

func BanFull(w http.ResponseWriter, r *http.Request) {
	err := banPlayer(r.URL, models.PlayerBanFull)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

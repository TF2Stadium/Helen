// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/bitly/go-simplejson"
)

type PlayerSummary struct {
	Avatar        string   `json:"avatar"`
	GameHours     int      `json:"gameHours"`
	ProfileURL    string   `json:"profileUrl"`
	LobbiesPlayed int      `json:"lobbiesPlayed"`
	SteamID       string   `json:"steamid"`
	Name          string   `json:"name"`
	Tags          []string `json:"tags"`
	Role          string   `json:"role"`
}

type Stats struct {
	Sixes      int `json:"playedSixesCount"`
	Highlander int `json:"playedHighlanderCount"`
	// Fours      int `json:"playedFoursCount"`
	// Ultiduo    int `json:"playedUltiduoCount"`
	// Bball      int `json:"playedBballCount"`
}

type PlayerProfile struct {
	Stats Stats `json:"stats"`

	CreatedAt int64  `json:"createdAt"`
	GameHours int    `json:"gameHours"`
	SteamID   string `json:"steamid"`
	Avatar    string `json:"avatar"`
	Name      string `json:"name"`
	ID        int    `json:"id"`
	Role      string `json:"role"`
}

func DecoratePlayerSettingsJson(settings []PlayerSetting) *simplejson.Json {
	json := simplejson.New()

	for _, obj := range settings {
		json.Set(obj.Key, obj.Value)
	}

	return json
}

func decoratePlayerTags(p *Player) []string {
	tags := []string{helpers.RoleNames[p.Role]}
	return tags
}

func DecoratePlayerProfileJson(p *Player) PlayerProfile {
	profile := PlayerProfile{}

	s := Stats{}
	s.Sixes = p.Stats.PlayedHighlanderCount
	s.Highlander = p.Stats.PlayedSixesCount
	profile.Stats = s

	// info
	profile.CreatedAt = p.CreatedAt.Unix()
	profile.GameHours = p.GameHours
	profile.SteamID = p.SteamId
	profile.Avatar = p.Avatar
	profile.Name = p.Name
	profile.Role = helpers.RoleNames[p.Role]

	// TODO ban info

	return profile
}

func DecoratePlayerSummary(p *Player) PlayerSummary {
	return PlayerSummary{
		Avatar:        p.Avatar,
		GameHours:     p.GameHours,
		ProfileURL:    p.Profileurl,
		LobbiesPlayed: p.Stats.PlayedHighlanderCount + p.Stats.PlayedSixesCount,
		SteamID:       p.SteamId,
		Name:          p.Name,
		Tags:          decoratePlayerTags(p),
		Role:          helpers.RoleNames[p.Role],
	}
}

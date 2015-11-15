// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/bitly/go-simplejson"
)

type PlayerSummary struct {
	Avatar        string   `json:"avatar,omitempty"`
	GameHours     int      `json:"gameHours,omitempty"`
	ProfileURL    string   `json:"profileUrl,omitempty"`
	LobbiesPlayed int      `json:"lobbiesPlayed,omitempty"`
	SteamID       string   `json:"steamid,omitempty"`
	Name          string   `json:"name,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Role          string   `json:"role,omitempty"`
}

type Stats struct {
	Sixes      int `json:"playedSixesCount"`
	Highlander int `json:"playedHighlanderCount"`
	// Fours      int `json:"playedFoursCount"`
	// Ultiduo    int `json:"playedUltiduoCount"`
	// Bball      int `json:"playedBballCount"`
}

type PlayerProfile struct {
	Stats Stats `json:"stats,omitempty"`

	CreatedAt int64  `json:"createdAt,omitempty"`
	GameHours int    `json:"gameHours,omitempty"`
	SteamID   string `json:"steamid,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Name      string `json:"name,omitempty"`
	ID        int    `json:"id,omitempty"`
	Role      string `json:"role,omitempty"`
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

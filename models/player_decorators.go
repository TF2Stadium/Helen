// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	db "github.com/TF2Stadium/Helen/database"
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

type PlayerProfile struct {
	Stats PlayerStats `json:"stats"`

	CreatedAt  int64  `json:"createdAt"`
	TwitchName string `json:"twitchName"`
	GameHours  int    `json:"gameHours"`
	SteamID    string `json:"steamid"`
	Avatar     string `json:"avatar"`
	Name       string `json:"name"`
	ID         int    `json:"id"`
	Role       string `json:"role"`
	//5 lobbies for now
	Lobbies []LobbyData `json:"lobbies"`
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
	db.DB.Preload("Stats").First(p, p.ID)
	profile := PlayerProfile{}
	alias, _ := p.GetSetting("siteAlias")

	p.Stats.Total = p.Stats.TotalLobbies()
	profile.Stats = p.Stats

	// info
	if alias.Value != "" {
		profile.Name = alias.Value
	}

	profile.CreatedAt = p.CreatedAt.Unix()
	profile.GameHours = p.GameHours
	profile.SteamID = p.SteamID
	profile.Avatar = p.Avatar
	profile.Name = p.Alias()
	profile.Role = helpers.RoleNames[p.Role]
	profile.TwitchName = p.TwitchName

	// TODO ban info
	var lobbies []*Lobby
	db.DB.Table("lobby_slots").Where("lobby_slots.player_id = ?", p.ID).Order("id desc").Limit("5").Find(&lobbies)
	for _, lobby := range lobbies {
		profile.Lobbies = append(profile.Lobbies, DecorateLobbyData(lobby, false))
	}
	return profile
}

func DecoratePlayerSummary(p *Player) PlayerSummary {
	db.DB.Preload("Stats").First(p, p.ID)
	summary := PlayerSummary{
		Avatar:        p.Avatar,
		GameHours:     p.GameHours,
		ProfileURL:    p.Profileurl,
		LobbiesPlayed: p.Stats.TotalLobbies(),
		SteamID:       p.SteamID,
		Name:          p.Alias(),
		Tags:          decoratePlayerTags(p),
		Role:          helpers.RoleNames[p.Role],
	}

	alias, _ := p.GetSetting("siteAlias")
	if alias.Value != "" {
		summary.Name = alias.Value
	}

	return summary
}

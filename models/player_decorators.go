// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"encoding/json"
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
)

func decoratePlayerTags(p *Player) []string {
	tags := []string{helpers.RoleNames[p.Role]}
	if p.IsStreaming {
		tags = append(tags, "twitch")
	}
	return tags
}

func (p *Player) setJSONFields(stats, lobbies, streaming, bans bool) {
	db.DB.Preload("Stats").First(p, p.ID)
	p.PlaceholderLobbiesPlayed = new(int)
	*p.PlaceholderLobbiesPlayed = p.Stats.TotalLobbies()

	if stats {
		p.Stats.Total = p.Stats.TotalLobbies()
		p.PlaceholderStats = &p.Stats
	}

	p.PlaceholderTags = new([]string)
	p.PlaceholderRoleStr = new(string)

	*p.PlaceholderRoleStr = helpers.RoleNames[p.Role]
	*p.PlaceholderTags = decoratePlayerTags(p)

	if lobbies {
		p.PlaceholderLobbies = new([]LobbyData)
		rows, err := db.DB.DB().Query("SELECT lobbies.ID FROM lobbies INNER JOIN lobby_slots ON lobbies.id = lobby_slots.lobby_id WHERE lobbies.match_ended = true AND lobby_slots.player_id = $1 ORDER BY lobbies.ID DESC LIMIT 5", p.ID)
		if err != nil {
			return
		}

		for rows.Next() {
			var id uint
			rows.Scan(&id)

			lobby, _ := GetLobbyByID(id)
			*p.PlaceholderLobbies = append(*p.PlaceholderLobbies, DecorateLobbyData(lobby, true))
		}
	}

	p.Name = p.Alias()
	if p.TwitchName != "" {
		if p.ExternalLinks == nil {
			p.ExternalLinks = make(map[string]*string)
		}

		twitchURL := "https://twitch.tv/" + p.TwitchName
		p.ExternalLinks["twitch"] = &twitchURL
	}

	p.setStreamingStatus()
	if bans {
		p.PlaceholderBans, _ = p.GetActiveBans()
	}
}

func (ban *PlayerBan) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type   string    `json:"type"`
		Until  time.Time `json:"until"`
		Reason string    `json:"reason"`
	}{ban.Type.String(), ban.Until, ban.Reason})
}

func (p *Player) SetPlayerProfile() {
	p.setJSONFields(true, true, true, true)
}

func (p *Player) SetPlayerSummary() {
	p.setJSONFields(false, false, false, false)
	p.ExternalLinks = nil
}

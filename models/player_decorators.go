// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
)

func decoratePlayerTags(p *Player) []string {
	tags := []string{helpers.RoleNames[p.Role]}
	return tags
}

func (p *Player) setJSONFields(stats, lobbies bool) {
	db.DB.Preload("Stats").First(p, p.ID)
	p.LobbiesPlayed = new(int)
	*p.LobbiesPlayed = p.Stats.TotalLobbies()

	if stats {
		p.PStats = &p.Stats
	}

	p.Tags = new([]string)
	p.RoleStr = new(string)

	*p.RoleStr = helpers.RoleNames[p.Role]
	*p.Tags = decoratePlayerTags(p)

	if lobbies {
		p.Lobbies = new([]LobbyData)
		rows, err := db.DB.DB().Query("SELECT lobbies.ID FROM lobbies INNER JOIN lobby_slots ON lobbies.id = lobby_slots.lobby_id WHERE lobbies.match_ended = true AND lobby_slots.player_id = $1 ORDER BY lobbies.ID DESC LIMIT 5", p.ID)
		if err != nil {
			return
		}

		for rows.Next() {
			var id uint
			rows.Scan(&id)

			lobby, _ := GetLobbyByID(id)
			*p.Lobbies = append(*p.Lobbies, DecorateLobbyData(lobby, true))
		}
	}

	p.Name = p.Alias()
}

func (p *Player) SetPlayerProfile() {
	p.setJSONFields(true, true)
}

var emptyStats = PlayerStats{}

func (p *Player) SetPlayerSummary() {
	p.setJSONFields(false, false)
	p.ExternalLinks = nil
	p.Stats = emptyStats
}

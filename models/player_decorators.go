// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/bitly/go-simplejson"
)

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

func DecoratePlayerProfileJson(p *Player) *simplejson.Json {
	j := simplejson.New()

	// stats
	s := simplejson.New()
	s.Set("playedHighlanderCount", p.Stats.PlayedHighlanderCount)
	s.Set("playedSixesCount", p.Stats.PlayedSixesCount)

	// info
	j.Set("createdAt", p.CreatedAt)
	j.Set("gameHours", p.GameHours)
	j.Set("steamid", p.SteamId)
	j.Set("avatar", p.Avatar)
	j.Set("stats", s)
	j.Set("name", p.Name)
	j.Set("id", p.ID)

	// TODO ban info

	return j
}

func DecoratePlayerSummaryJson(p *Player) *simplejson.Json {
	j := simplejson.New()

	j.Set("id", p.ID)
	j.Set("avatar", p.Avatar)
	j.Set("gameHours", p.GameHours)
	j.Set("profileUrl", p.Profileurl)
	j.Set("lobbiesPlayed", p.Stats.PlayedHighlanderCount+p.Stats.PlayedSixesCount)
	j.Set("steamid", p.SteamId)
	j.Set("name", p.Name)
	j.Set("tags", decoratePlayerTags(p))
	j.Set("role", helpers.RoleNames[p.Role])

	return j
}

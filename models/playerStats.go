// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"github.com/TF2Stadium/Helen/database"
)

type PlayerStats struct {
	ID uint `json:"-"`

	Total                 int `sql:"-" json:"lobbiesPlayed"`
	PlayedSixesCount      int `sql:"played_sixes_count",default:"0" json:"playedSixesCount"`
	PlayedHighlanderCount int `sql:"played_highlander_count",default:"0" json:"playedHighlanderCount"`
	PlayedFoursCount      int `sql:"played_fours_count",json:"playedFoursCount" `
	PlayedUltiduoCount    int `sql:"played_ultiduo_count",json:"playedUltiduoCount"`
	PlayedBballCount      int `sql:"played_bball_count",json:"playedBballCount"`

	Substitutes int `json:"substitutes"`
}

func NewPlayerStats() PlayerStats {
	stats := PlayerStats{}

	return stats
}

func (ps *PlayerStats) TotalLobbies() int {
	return ps.PlayedSixesCount + ps.PlayedHighlanderCount + ps.PlayedFoursCount + ps.PlayedUltiduoCount + ps.PlayedBballCount
}

func (ps *PlayerStats) PlayedCountIncrease(lt LobbyType) {
	switch lt {
	case LobbyTypeSixes:
		ps.PlayedSixesCount++
	case LobbyTypeHighlander:
		ps.PlayedHighlanderCount++
	case LobbyTypeFours:
		ps.PlayedFoursCount++
	case LobbyTypeBball:
		ps.PlayedBballCount++
	case LobbyTypeUltiduo:
		ps.PlayedUltiduoCount++
	}
	database.DB.Save(ps)
}

func (ps *PlayerStats) IncreaseSubCount() {
	ps.Substitutes++
	database.DB.Save(ps)
}

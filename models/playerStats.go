// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

type PlayerStats struct {
	ID                    uint `json:"-"`
	PlayedSixesCount      int  `sql:"played_sixes_count",default:"0"`
	PlayedHighlanderCount int  `sql:"played_highlander_count",default:"0"`
	PlayedFoursCount      int  `sql:"played_fours_count",json:"playedFoursCount"`
	PlayedUltiduoCount    int  `sql:"played_ultiduo_count",json:"playedUltiduoCount"`
	PlayedBballCount      int  `sql:"played_bball_count",json:"playedBballCount"`
}

func NewPlayerStats() PlayerStats {
	stats := PlayerStats{}

	return stats
}

func (ps *PlayerStats) PlayedCountIncrease(lt LobbyType) {
	switch lt {
	case LobbyTypeSixes:
		ps.PlayedSixesCount += 1
	case LobbyTypeHighlander:
		ps.PlayedHighlanderCount += 1
	case LobbyTypeFours:
		ps.PlayedFoursCount += 1
	case LobbyTypeBball:
		ps.PlayedBballCount += 1
	case LobbyTypeUltiduo:
		ps.PlayedUltiduoCount += 1
	}
}

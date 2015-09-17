// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

type PlayerStats struct {
	ID                    uint
	PlayedSixesCount      int `sql:"played_sixes_count",default:"0"`
	PlayedHighlanderCount int `sql:"played_highlander_count",default:"0"`
}

func NewPlayerStats() PlayerStats {
	stats := PlayerStats{}

	return stats
}

func (ps *PlayerStats) PlayedCountSet(lt LobbyType, value int) {
	switch lt {
	case LobbyTypeSixes:
		ps.PlayedSixesCount = value
	case LobbyTypeHighlander:
		ps.PlayedHighlanderCount = value
	}
}

func (ps *PlayerStats) PlayedCountGet(lt LobbyType) int {
	switch lt {
	case LobbyTypeSixes:
		return ps.PlayedSixesCount
	case LobbyTypeHighlander:
		return ps.PlayedHighlanderCount
	}
	return 0
}

func (ps *PlayerStats) PlayedCountIncrease(lt LobbyType) {
	switch lt {
	case LobbyTypeSixes:
		ps.PlayedSixesCount += 1
	case LobbyTypeHighlander:
		ps.PlayedHighlanderCount += 1
	}
}

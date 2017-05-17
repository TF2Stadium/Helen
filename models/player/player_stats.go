// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package player

import (
	"time"

	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models/lobby/format"
)

type PlayerStats struct {
	ID uint `json:"-"`

	Total                 int `sql:"-" json:"lobbiesPlayed"`
	PlayedSixesCount      int `sql:"played_sixes_count",default:"0" json:"playedSixesCount"`
	PlayedHighlanderCount int `sql:"played_highlander_count",default:"0" json:"playedHighlanderCount"`
	PlayedFoursCount      int `sql:"played_fours_count",json:"playedFoursCount" `
	PlayedUltiduoCount    int `sql:"played_ultiduo_count",json:"playedUltiduoCount"`
	PlayedBballCount      int `sql:"played_bball_count",json:"playedBballCount"`
	PlayedProlanderCount  int `sql:"played_prolander_count",json:"playedProlanderCount"`

	Scout         int           `json:"scout"`
	ScoutHours    time.Duration `json:"scoutHours"`
	Soldier       int           `json:"soldier"`
	SoldierHours  time.Duration `json:"soldierHours"`
	Pyro          int           `json:"pyro"`
	PyroHours     time.Duration `json:"pyroHours"`
	Engineer      int           `json:"engineer"`
	EngineerHours time.Duration `json:"engineerHours"`
	Heavy         int           `json:"heavy"`
	HeavyHours    time.Duration `json:"heavyHours"`
	Demoman       int           `json:"demoman"`
	DemoHours     time.Duration `json:"demomanHours"`
	Sniper        int           `json:"sniper"`
	SniperHours   time.Duration `json:"sniperHours"`
	Medic         int           `json:"medic"`
	MedicHours    time.Duration `json:"medicHours"`
	Spy           int           `json:"spy"`
	SpyHours      time.Duration `json:"spyHours"`

	Substitutes int `json:"substitutes"`
}

func NewStats() PlayerStats {
	stats := PlayerStats{}

	return stats
}

func (ps *PlayerStats) Save() {
	database.DB.Save(ps)
}

func (ps *PlayerStats) TotalLobbies() int {
	return ps.PlayedSixesCount + ps.PlayedHighlanderCount + ps.PlayedFoursCount + ps.PlayedUltiduoCount + ps.PlayedBballCount
}

func (ps *PlayerStats) PlayedCountIncrease(lt format.Format) {
	switch lt {
	case format.Sixes:
		ps.PlayedSixesCount++
	case format.Highlander:
		ps.PlayedHighlanderCount++
	case format.Fours:
		ps.PlayedFoursCount++
	case format.Bball:
		ps.PlayedBballCount++
	case format.Ultiduo:
		ps.PlayedUltiduoCount++
	case format.Prolander:
		ps.PlayedProlanderCount++
	}
	database.DB.Save(ps)
}

func (ps *PlayerStats) IncreaseSubCount() {
	ps.Substitutes++
	database.DB.Save(ps)
}

func (ps *PlayerStats) IncreaseClassCount(f format.Format, slot int) {
	_, class, _ := format.GetSlotTeamClass(f, slot)
	switch class {
	case "scout", "scout1", "scout2":
		ps.Scout++
	case "roamer", "pocket", "soldier", "soldier1", "soldier2":
		ps.Soldier++
	case "pyro":
		ps.Pyro++
	case "engineer":
		ps.Engineer++
	case "heavy":
		ps.Heavy++
	case "demoman":
		ps.Demoman++
	case "sniper":
		ps.Sniper++
	case "medic":
		ps.Medic++
	case "spy":
		ps.Spy++
	}
	database.DB.Save(ps)
}

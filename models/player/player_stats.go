// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package player

import (
	"time"

	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models/lobby/format"
)

type Stats struct {
	ID uint `json:"-"`

	Total                 int `sql:"-" json:"lobbiesPlayed"`
	PlayedSixesCount      int `sql:"played_sixes_count",default:"0" json:"playedSixesCount"`
	PlayedHighlanderCount int `sql:"played_highlander_count",default:"0" json:"playedHighlanderCount"`
	PlayedFoursCount      int `sql:"played_fours_count",json:"playedFoursCount" `
	PlayedUltiduoCount    int `sql:"played_ultiduo_count",json:"playedUltiduoCount"`
	PlayedBballCount      int `sql:"played_bball_count",json:"playedBballCount"`

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
	Headshots   int `json:"headshots"`
	Airshots    int `json:"airshots"`
}

func NewStats() Stats {
	stats := Stats{}

	return stats
}

func (ps *Stats) TotalLobbies() int {
	return ps.PlayedSixesCount + ps.PlayedHighlanderCount + ps.PlayedFoursCount + ps.PlayedUltiduoCount + ps.PlayedBballCount
}

func (ps *Stats) PlayedCountIncrease(lt format.Format) {
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
	}
	database.DB.Save(ps)
}

func (ps *Stats) IncreaseSubCount() {
	ps.Substitutes++
	database.DB.Save(ps)
}

func (ps *Stats) IncreaseClassCount(f format.Format, slot int) {
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

type Report struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time

	PlayerID uint
	LobbyID  uint
	Type     ReportType
}

type ReportType int

const (
	Substitute ReportType = iota //!sub
	Vote                         //!repped by other players
	RageQuit                     //rage quit
)

func (player *Player) NewReport(rtype ReportType, lobbyid uint) {
	var count int

	last := time.Now().Add(-30 * time.Minute)
	database.DB.Model(&Report{}).Where("player_id = ? AND created_at > ? AND type = ?", player.ID, last, rtype).Count(&count)

	switch rtype {
	case Substitute:
		if count == 2 {
			player.BanUntil(time.Now().Add(30*time.Minute), BanJoin, "For !subbing twice in the last 30 minutes")
		}
	case Vote:
		if count != 0 {
			player.BanUntil(time.Now().Add(30*time.Minute), BanJoin, "For getting !repped from a lobby in the last 30 minutes")
		}
	case RageQuit:
		if count != 0 {
			player.BanUntil(time.Now().Add(30*time.Minute), BanJoin, "For ragequitting a lobby in the last 30 minutes")
		}

	}

	r := &Report{
		LobbyID:  lobbyid,
		PlayerID: player.ID,
		Type:     rtype,
	}
	database.DB.Save(r)
}

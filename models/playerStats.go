// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import "github.com/TF2Stadium/Helen/database"

type PlayerStats struct {
	ID uint `json:"-"`

	Total                 int `sql:"-" json:"lobbiesPlayed"`
	PlayedSixesCount      int `sql:"played_sixes_count",default:"0" json:"playedSixesCount"`
	PlayedHighlanderCount int `sql:"played_highlander_count",default:"0" json:"playedHighlanderCount"`
	PlayedFoursCount      int `sql:"played_fours_count",json:"playedFoursCount" `
	PlayedUltiduoCount    int `sql:"played_ultiduo_count",json:"playedUltiduoCount"`
	PlayedBballCount      int `sql:"played_bball_count",json:"playedBballCount"`

	Scout    int `json:"scout"`
	Soldier  int `json:"soldier"`
	Pyro     int `json:"pyro"`
	Engineer int `json:"engineer"`
	Heavy    int `json:"heavy"`
	Demoman  int `json:"demoman"`
	Sniper   int `json:"sniper"`
	Medic    int `json:"medic"`
	Spy      int `json:"spy"`

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

func (ps *PlayerStats) IncreaseClassCount(lobby *Lobby, slot int) {
	_, class, _ := LobbyGetSlotInfoString(lobby.Type, slot)
	classes := map[string]*int{
		"scout":    &ps.Scout,
		"scout1":   &ps.Scout,
		"scout2":   &ps.Scout,
		"roamer":   &ps.Soldier,
		"pocket":   &ps.Soldier,
		"soldier":  &ps.Soldier,
		"soldier1": &ps.Soldier,
		"soldier2": &ps.Soldier,
		"pyro":     &ps.Pyro,
		"engineer": &ps.Engineer,
		"heavy":    &ps.Heavy,
		"demoman":  &ps.Demoman,
		"sniper":   &ps.Sniper,
		"medic":    &ps.Medic,
		"spy":      &ps.Spy,
	}

	*(classes[class])++
	database.DB.Save(ps)
}

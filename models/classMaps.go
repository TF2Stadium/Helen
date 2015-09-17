// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import "github.com/TF2Stadium/Helen/helpers"

var teamMap = map[string]int{"red": 0, "blu": 1}
var sixesClassMap = map[string]int{
	"scout1":  0,
	"scout2":  1,
	"roamer":  2,
	"pocket":  3,
	"demoman": 4,
	"medic":   5,
}

var sixesClassList = []string{"scout1", "scout2", "roamer", "pocket", "demoman", "medic"}

var hlClassMap = map[string]int{
	"scout":    0,
	"soldier":  1,
	"pyro":     2,
	"demoman":  3,
	"heavy":    4,
	"engineer": 5,
	"medic":    6,
	"sniper":   7,
	"spy":      8,
}

var hlClassList = []string{"scout", "soldier", "pyro", "demoman", "heavy", "engineer", "medic", "sniper", "spy"}

func LobbyGetPlayerSlot(lobbytype LobbyType, teamStr string, classStr string) (int, *helpers.TPError) {
	team, ok := teamMap[teamStr]
	if !ok {
		return -1, helpers.NewTPError("Invalid team", -1)
	}

	var classMap map[string]int
	switch lobbytype {
	case LobbyTypeHighlander:
		classMap = hlClassMap
	case LobbyTypeSixes:
		classMap = sixesClassMap
	}

	class, ok := classMap[classStr]

	if !ok {
		return -1, helpers.NewTPError("Invalid class", -1)
	}

	return team*len(classMap) + class, nil
}

func LobbyFormatClassMap(format LobbyType) map[string]int {
	if format == LobbyTypeHighlander {
		return hlClassMap

	}
	return sixesClassMap
}

func LobbyFormatClassList(format LobbyType) []string {
	if format == LobbyTypeHighlander {
		return hlClassList

	}
	return sixesClassList
}

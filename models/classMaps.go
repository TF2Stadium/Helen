// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"github.com/TF2Stadium/Helen/helpers"
)

var teamMap = map[string]int{"red": 0, "blu": 1}
var teamList = []string{"red", "blu"}

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

var debugClassMap = map[string]int{
	"scout": 0,
}
var debugClassList = []string{"scout"}

var bballClassMap = map[string]int{
	"soldier1": 0,
	"soldier2": 1,
}
var bballClassList = []string{"soldier1", "soldier2"}

var ultiduoClassMap = map[string]int{
	"soldier": 0,
	"medic":   1,
}
var ultiduoClassList = []string{"soldier", "medic"}

var foursClassMap = map[string]int{
	"scout":   0,
	"soldier": 1,
	"demoman": 2,
	"medic":   3,
}
var foursClassList = []string{"scout", "soldier", "demoman", "medic"}

var TypeClassMap = map[LobbyType]map[string]int{
	LobbyTypeHighlander: hlClassMap,
	LobbyTypeSixes:      sixesClassMap,
	LobbyTypeFours:      foursClassMap,
	LobbyTypeUltiduo:    ultiduoClassMap,
	LobbyTypeBball:      bballClassMap,
	LobbyTypeDebug:      debugClassMap,
}

var TypeClassList = map[LobbyType][]string{
	LobbyTypeHighlander: hlClassList,
	LobbyTypeSixes:      sixesClassList,
	LobbyTypeFours:      foursClassList,
	LobbyTypeUltiduo:    ultiduoClassList,
	LobbyTypeBball:      bballClassList,
	LobbyTypeDebug:      debugClassList,
}

var NumberOfClassesMap = map[LobbyType]int{
	LobbyTypeHighlander: 9,
	LobbyTypeSixes:      6,
	LobbyTypeFours:      4,
	LobbyTypeUltiduo:    2,
	LobbyTypeBball:      2,
	LobbyTypeDebug:      1,
}

func LobbyGetPlayerSlot(lobbytype LobbyType, teamStr string, classStr string) (int, *helpers.TPError) {
	team, ok := teamMap[teamStr]
	if !ok {
		return -1, helpers.NewTPError("Invalid team", -1)
	}

	class, ok := TypeClassMap[lobbytype][classStr]
	if !ok {
		return -1, helpers.NewTPError("Invalid class", -1)
	}

	return team*NumberOfClassesMap[lobbytype] + class, nil
}

func LobbyGetSlotInfoString(lobbytype LobbyType, slot int) (string, string, *helpers.TPError) {
	classList := TypeClassList[lobbytype]

	team, class, err := LobbyGetSlotInfo(lobbytype, slot)
	if err == nil {
		return teamList[team], classList[class], nil
	}
	return "", "", err
}

func LobbyGetSlotInfo(lobbytype LobbyType, slot int) (int, int, *helpers.TPError) {
	classList := TypeClassList[lobbytype]

	if slot < len(classList) {
		return 0, slot, nil
	} else if slot < 2*len(classList) {
		return 1, slot - len(classList), nil
	} else {
		return 0, 0, helpers.NewTPError("Invalid slot", -1)
	}
}

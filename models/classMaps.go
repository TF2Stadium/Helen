// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"github.com/TF2Stadium/Helen/helpers"
	"strings"
)

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
	"demo":    2,
	"medic":   3,
}
var foursClassList = []string{"scout", "soldier", "demo", "medic"}

var TypeClassMap = map[LobbyType]map[string]int{
	LobbyTypeHighlander: hlClassMap,
	LobbyTypeSixes:      sixesClassMap,
	LobbyTypeDebug:      debugClassMap,
}

var typeClassList = map[LobbyType][]string{
	LobbyTypeHighlander: hlClassList,
	LobbyTypeSixes:      sixesClassList,
	LobbyTypeDebug:      debugClassList,
	LobbyTypeBball:      bballClassList,
	LobbyTypeFours:      foursClassList,
}

func TypeClassList(l LobbyType, mapname string) []string {
	list := typeClassList[l]
	if strings.HasPrefix(mapname, "ultiduo") || strings.HasPrefix(mapname, "koth_ultiduo") {
		list = ultiduoClassList
	}
	return list
}

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
	case LobbyTypeDebug:
		classMap = debugClassMap
	case LobbyTypeFours:
		classMap = foursClassMap
	case LobbyTypeUltiduo:
		classMap = bballClassMap
		if classStr == "soldier" || classStr == "medic" {
			classMap = ultiduoClassMap
		}
	}

	class, ok := classMap[classStr]

	if !ok {
		return -1, helpers.NewTPError("Invalid class", -1)
	}

	return team*len(classMap) + class, nil
}

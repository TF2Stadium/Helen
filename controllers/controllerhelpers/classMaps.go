package controllerhelpers

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
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

func GetPlayerSlot(lobbytype models.LobbyType, teamStr string, classStr string) (int, *helpers.TPError) {
	team, ok := teamMap[teamStr]
	if !ok {
		return -1, helpers.NewTPError("Invalid team", -1)
	}

	var classMap map[string]int
	switch lobbytype {
	case models.LobbyTypeHighlander:
		classMap = hlClassMap
	case models.LobbyTypeSixes:
		classMap = sixesClassMap
	}

	class, ok := classMap[classStr]

	if !ok {
		return -1, helpers.NewTPError("Invalid class", -1)
	}

	return team*len(classMap) + class, nil
}

func FormatClassMap(format models.LobbyType) map[string]int {
	if format == models.LobbyTypeHighlander {
		return hlClassMap

	}
	return sixesClassMap
}

func FormatClassList(format models.LobbyType) []string {
	if format == models.LobbyTypeHighlander {
		return hlClassList

	}
	return sixesClassList
}

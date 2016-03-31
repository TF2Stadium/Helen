package format

import "errors"

type Format int

const (
	Sixes      Format = iota
	Highlander        // lol
	Fours
	Ultiduo
	Bball
	Debug
)

var (
	teamMap  = map[string]int{"red": 0, "blu": 1}
	teamList = []string{"red", "blu"}

	sixesClassMap = map[string]int{
		"scout1":  0,
		"scout2":  1,
		"roamer":  2,
		"pocket":  3,
		"demoman": 4,
		"medic":   5,
	}
	sixesClassList = []string{"scout1", "scout2", "roamer", "pocket", "demoman", "medic"}

	hlClassMap = map[string]int{
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
	hlClassList = []string{
		"scout",
		"soldier",
		"pyro",
		"demoman",
		"heavy",
		"engineer",
		"medic",
		"sniper",
		"spy"}

	debugClassMap = map[string]int{
		"scout": 0,
	}
	debugClassList = []string{"scout"}

	bballClassMap = map[string]int{
		"soldier1": 0,
		"soldier2": 1,
	}
	bballClassList = []string{"soldier1", "soldier2"}

	ultiduoClassMap = map[string]int{
		"soldier": 0,
		"medic":   1,
	}
	ultiduoClassList = []string{"soldier", "medic"}

	foursClassMap = map[string]int{
		"scout":   0,
		"soldier": 1,
		"demoman": 2,
		"medic":   3,
	}
	foursClassList = []string{"scout", "soldier", "demoman", "medic"}

	typeClassMap = map[Format]map[string]int{
		Highlander: hlClassMap,
		Sixes:      sixesClassMap,
		Fours:      foursClassMap,
		Ultiduo:    ultiduoClassMap,
		Bball:      bballClassMap,
		Debug:      debugClassMap,
	}

	typeClassList = map[Format][]string{
		Highlander: hlClassList,
		Sixes:      sixesClassList,
		Fours:      foursClassList,
		Ultiduo:    ultiduoClassList,
		Bball:      bballClassList,
		Debug:      debugClassList,
	}

	NumberOfClassesMap = map[Format]int{
		Highlander: 9,
		Sixes:      6,
		Fours:      4,
		Ultiduo:    2,
		Bball:      2,
		Debug:      1,
	}
)

//GetSlot returns the slot number for given team, class strings and the
//lobby format
func GetSlot(lobbytype Format, teamStr string, classStr string) (int, error) {
	team, ok := teamMap[teamStr]
	if !ok {
		return -1, errors.New("Invalid team")
	}

	class, ok := typeClassMap[lobbytype][classStr]
	if !ok {
		return -1, errors.New("Invalid class")
	}

	return team*NumberOfClassesMap[lobbytype] + class, nil
}

//GetSlotTeamClass returns the team and class strings for a given slot number
func GetSlotTeamClass(lobbytype Format, slot int) (team, class string, err error) {
	classList := typeClassList[lobbytype]

	teamI, classI, err := getSlotNums(lobbytype, slot)
	if err == nil {
		team, class, err = teamList[teamI], classList[classI], nil
	}
	return
}

//given a slot number, returns the numbers for the
//slot's class and team for the given format
func getSlotNums(lobbytype Format, slot int) (int, int, error) {
	classList := typeClassList[lobbytype]

	if slot < len(classList) {
		return 0, slot, nil
	} else if slot < 2*len(classList) {
		return 1, slot - len(classList), nil
	} else {
		return 0, 0, errors.New("Invalid slot")
	}
}

func GetClasses(format Format) []string {
	return typeClassList[format]
}

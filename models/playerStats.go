package models

import (
	"strconv"
	"strings"
)

type LobbiesPlayed struct {
	Data  string
	Total map[LobbyType]int
}

// parse string from Data
func (lp *LobbiesPlayed) Parse() {
	list := strings.Split(lp.Data, ",")

	for i, v := range list {
		lp.Total[LobbyType(i)], _ = strconv.Atoi(v)
	}
}

// build string from Total to Data and returns it
func (lp *LobbiesPlayed) String() string {
	lp.Data = ""

	// build string "5,2,9..." to lp.Data
	for i := 0; i < len(lp.Total); i++ {
		t := strconv.Itoa(lp.Total[LobbyType(i)])
		lp.Data += (t + ",")
	}

	ds := len(lp.Data)

	// remove last comma
	if lp.Data[ds-1:ds] == "," {
		lp.Data = lp.Data[:ds-1]
	}

	return lp.Data
}

func (lp *LobbiesPlayed) Set(lt LobbyType, value int) {
	lp.Total[lt] = value

	_ = lp.String()
}

func (lp *LobbiesPlayed) Get(lt LobbyType) int {
	return lp.Total[lt]
}

func (lp *LobbiesPlayed) Increase(lt LobbyType) {
	lp.Total[lt] += 1

	_ = lp.String()
}

type PlayerStats struct {
	LobbiesPlayed *LobbiesPlayed
}

func NewPlayerStats() *PlayerStats {
	stats := new(PlayerStats)
	stats.LobbiesPlayed = new(LobbiesPlayed)
	stats.LobbiesPlayed.Total = make(map[LobbyType]int)

	return stats
}

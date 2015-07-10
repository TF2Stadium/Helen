package models

import (
	"container/list"
	"fmt"
	"net/http"
)

//Given Lobby IDs are unique, we'll use them for mumble channel names
type Lobby struct {
	id      int        //Lobby id
	mapName string     //map name
	team    [][]string /* RED - team[0], BLU - team[1]
	 * players are identified by their DB id */
	server    string //server address, with port
	rconpwd   string //password to server's rcon
	whitelist int    //whitelist.tf ID
}

var dbLobbyMap = make(map[string]*Lobby) //maps mongoDB id --> lobby
var lobbyMap = make(map[int]*Lobby)      //maps lobby id --> lobby

//id should be maintained in the main loop
func NewLobby(mapName string, players int, server string, rconpwd string, id int,
	whitelist int) *Lobby {
	lobby := &Lobby{
		id:        id,
		mapName:   mapName,
		team:      make([][]string, 2),
		server:    server,
		rconpwd:   rconpwd,
		whitelist: whilelist,
	}

	lobby.team[0] = make([]*Player, players)
	lobby.team[1] = make([]*Player, players)

	return lobby
}

func CloseLobby(id int) {
	delete(lobbyMap, id)
}

//Add player to lobby
func (lobby *Lobby) AddPlayer(id string, team int, slot int) error {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */
	if _, exists := dbLobbyMap[id]; exists {
		//player is already in a lobby
		return 1
	}

	if lobby.team[team][slot].id != "" {
		//slot has been filled
		return 2
	}

	dbLobbyMap[id] = lobby
	lobby.team[team][slot] = id

	//Check if all slots have been filled
	for _, team := range lobby.team {
		for _, player := range team {
			if player.id == "" { //Slots haven't been filled
				return nil
			}
		}
	}
	/* Slots have been filled. Ask frontend to create the "Ready" button.
	 */
	return nil
}

//Remove player from lobby
func (lobby *Lobby) Remove(id string) {
	delete(dbLobbyMap, id)
}

//Return an array of all lobbies
func GetLobbyList(open bool) {
	arr := make([]*Lobby, len(lobbyMap))

	i := 0
	for _, lobby := range lobbyID {
		arr[i] = lobby
		i++
	}

	return arr
}

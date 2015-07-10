package models

import (
	"container/list"
	"fmt"
	"net/http"
)

//Given Lobby IDs are unique, we'll use them for mumble channel names
type Lobby struct {
	id        int         //Lobby id
	mapName   string      //map name
	team      [][]*Player //RED - team[0], BLU - team[1]
	server    string      //server address, with port
	rconpwd   string      //password to server's rcon
	whitelist int         //whitelist.tf ID
}

var steamPlayerMap = make(map[string]*Player) //maps steamid --> player
var steamLobbyMap = make(map[string]*Lobby)   //maps steamid --> lobby
var LobbyMap = make(map[int]*Lobby)           //maps looby id --> lobby
var lobList = list.New()                      //list of all lobbies

//id should be maintained in the main loop
func NewLobby(mapName string, players int, server string, rconpwd string, id int,
	whitelist int) *Lobby {
	lobby := &Lobby{
		id:        id,
		mapName:   mapName,
		team:      make([][]*Player, 2),
		server:    server,
		rconpwd:   rconpwd,
		whitelist: whilelist,
	}

	lobby.team[0] = make([]*Player, players)
	lobby.team[1] = make([]*Player, players)

	return lobby
}

//Add player to lobby
func (lobby *Lobby) Add(steamid string, name string, team int, slot int) error {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */
	if _, exists := steamLobbyMap[steamid]; exists {
		//return error, player is already in a lobby
	}

	if lobby.team[team][slot].SteamId != "" {
		//return error, slot has been filled
	}

	steamLobbyMap[steamid] = lobby
	player := &Player{SteamId: steamid, Name: name}
	steamPlayerMap[steamid] = player
	lobby.team[team][slot] = player

	//Check if all slots have been filled
	for _, team := range lobby.team {
		for _, player := range team {
			if player.SteamId == "" { //Slots haven't been filled
				return nil
			}
		}
	}
	/* Slots have been filled. Ask frontend to create the "Ready" button.
	 */
	return nil
}

//Remove player from lobby
func (lobby *Lobby) Remove(steamid string) {
	delete(steamLobbyMap, steamid)
	steamPlayerMap[steamid] = &Player{SteamId: "", Name: "id"}
	delete(steamPlayerMap, steamid)
}

package models

import "github.com/TeamPlayTF/Server/helpers"

//Given Lobby IDs are unique, we'll use them for mumble channel names
type Lobby struct {
	id        int         //Lobby id
	mapName   string      //map name
	team      [][]*Player // RED - team[0], BLU - team[1]
	server    string      //server address, with port
	rconpwd   string      //password to server's rcon
	whitelist int         //whitelist.tf ID
}

var dbLobbyMap = make(map[string]*Lobby) //maps mongoDB id --> lobby
var lobbyMap = make(map[int]*Lobby)      //maps lobby id --> lobby

//id should be maintained in the main loop
func NewLobby(mapName string, players int, server string, rconpwd string, id int, whitelist int) *Lobby {
	lobby := &Lobby{
		id:        id,
		mapName:   mapName,
		team:      make([][]*Player, 2),
		server:    server,
		rconpwd:   rconpwd,
		whitelist: whitelist,
	}

	lobby.team[0] = make([]*Player, players)
	lobby.team[1] = make([]*Player, players)

	return lobby
}

//Close lobby, given the lobby id
func CloseLobby(id int) {
	delete(lobbyMap, id)
}

//Add player to lobby
func (lobby *Lobby) AddPlayer(player *Player, team int, slot int) error {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */
	if _, exists := dbLobbyMap[player.Id.Hex()]; exists {
		//player is already in a lobby
		return helpers.NewError("Player is already in a lobby", 1)
	}

	if lobby.team[team][slot].Id != "" {
		//slot has been filled
		return helpers.NewError("This slot has been filled.", 2)
	}

	dbLobbyMap[player.Id.Hex()] = lobby
	lobby.team[team][slot] = player

	//Check if all slots have been filled
	for _, team := range lobby.team {
		for _, player := range team {
			if player.Id == "" { //Slots haven't been filled
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
func GetLobbyList(open bool) []*Lobby {
	var arr []*Lobby

	for _, val := range lobbyMap {
		arr = append(arr, val)
	}

	return arr
}

func GetLobbyDetails(id int) *Lobby {
	return lobbyMap[id]
}

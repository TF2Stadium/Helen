package server
import "container/list"

type Player struct {
	steamid string //Players steam ID
	name string //Player name
	ready bool
}

//Given Lobby IDs are unique, we'll use them for mumble channel names
type Lobby struct {
	id int //Lobby id
	map_name string // map name
	team [][]*Player //RED - team[0], BLU - team[1]
	server string //server address, with port
	rconpwd string  //password to server's rcon
}

var steamPlayerMap = make(map[string]*Player) //maps steamid --> player
var steamLobbyMap = make(map[string]*Lobby) //maps steamid --> lobby
var LobbyMap = make(map[int]*Lobby)     //maps looby id --> lobby
var lobList = list.New() // list of all lobbies

//id should be maintained in the main loop
func New(map_name string, players int, server string, rconpwd string, id int) *Lobby {
	lobby := &Lobby {
		id: id,
		map_name: map_name,
		team: make([][]*Player, 2),
		server: server,
		rconpwd: rconpwd,
	}
	
	lobby.team[0] = make([]*Player, players)
	lobby.team[1] = make([]*Player, players)	

	return lobby
}

//Add player to lobby
func (lobby *Lobby) Add(steamid string, name string, team int, slot int) error{
	/* Possible errors while joining
         * Slot has been filled
         * Player has already joined a lobby
         * anything else?
         */
	if _, exists := steamLobbyMap[steamid]; exists {
		//return error, player is already in a lobby
	}
	
	if (lobby.team[team][slot].steamid != "") {
		//return error, slot has been filled
	}

	steamLobbyMap[steamid] = lobby
	player := &Player{steamid: steamid, name: name, ready: false}
	steamPlayerMap[steamid] = player
	lobby.team[team][slot] = player

	//Check if all slots have been filled
	for _, team := range lobby.team {
		for _, player := range team {
			if (player.steamid == "") { //Slots haven't been filled
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
	steamPlayerMap[steamid] = &Player {steamid: "", name: "id", ready: false}
	delete(steamPlayerMap, steamid)
}

/* Loops through all players of both teams in lobby, returns true if all players
 * ready, else false.
*/
func (lobby *Lobby) AllReady() bool {
	for _, team := range lobby.team {
		for _, player := range team {
			if !player.ready {
				return false
			}
		}
	}
	return true
}

func Ready(steamid string) {
	steamPlayerMap[steamid].ready = true
	if steamLobbyMap[steamid].AllReady() { //Everybody's ready
		/* Create mumble channel, use RCON interface to start match
		 * Ask frontend to create join buttons.
                 */
	}
}

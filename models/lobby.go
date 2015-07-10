package models

import (
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
)

//Given Lobby IDs are unique, we'll use them for mumble channel names
type Lobby struct {
	id      int         //Lobby id
	mapName string      // map name
	team    [][]*Player //RED - team[0], BLU - team[1]
	server  string      //server address, with port
	rconpwd string      //password to server's rcon
}

//Response structure
//github.com/TeamPlayTF/Specifications/blob/master/Communication.md#response-format
type Response struct {
	Successful bool        `json:"successful"` //true if operation was successful
	Data       interface{} `json:"data"`       //response message, if any
	Code       int         `json: "code"`      //errcode, if sucessful == false
}

var steamPlayerMap = make(map[string]*Player) //maps steamid --> player
var steamLobbyMap = make(map[string]*Lobby)   //maps steamid --> lobby
var LobbyMap = make(map[int]*Lobby)           //maps looby id --> lobby
var lobList = list.New()                      //list of all lobbies

func SendJSON(w http.ResponseWriter, j string) {
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, j)
}

func SendError(w http.ResponseWriter, code int, message string) string {
	r := &Response{
		Successful: false,
		Data:       data,
		Code:       code,
	}
	j, _ := json.Marshall(r)
	sendJSON(w, j)
	return string(j)
}

func SendSuccess(w http.ResponseWriter, data interface{}) string {
	r := &Response{
		Successful: true,
		Data:       data,
		Code:       -1,
	}
	j, _ := json.Marshall(r)
	sendJSON(w, j)
	return string(j)
}

//id should be maintained in the main loop
func NewLobby(mapName string, players int, server string, rconpwd string, id int) *Lobby {
	lobby := &Lobby{
		id:      id,
		mapName: mapName,
		team:    make([][]*Player, 2),
		server:  server,
		rconpwd: rconpwd,
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

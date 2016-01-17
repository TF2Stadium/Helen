//Package sessions provides functions to help maintain consistency
//across multiple websocket connections from a single player,
//when the player has multiple tabs/windows open (since each tab opens a new websocket connection)
package sessions

import (
	"sync"

	"github.com/TF2Stadium/wsevent"
)

var (
	mapMu = new(sync.RWMutex)
	//steamid -> client array, since players can have multiple tabs open
	steamIDSockets = make(map[string][]*wsevent.Client)
	//socketid -> id of lobby the socket is spectating
	socketSpectating = make(map[string]uint)
)

//AddSocket adds so to the list of sockets connected from steamid
func AddSocket(steamid string, so *wsevent.Client) {
	mapMu.Lock()
	defer mapMu.Unlock()

	steamIDSockets[steamid] = append(steamIDSockets[steamid], so)
}

//RemoveSocket removes so from the list of sockets connected from steamid
func RemoveSocket(sessionID, steamID string) {
	mapMu.Lock()
	defer mapMu.Unlock()

	clients := steamIDSockets[steamID]
	for i, socket := range clients {
		if socket.Id() == sessionID {
			clients[i] = clients[len(clients)-1]
			clients[len(clients)-1] = nil
			clients = clients[:len(clients)-1]
			break
		}
	}

	steamIDSockets[steamID] = clients

	if len(clients) == 0 {
		delete(steamIDSockets, steamID)
	}
}

//GetSockets returns a list of sockets connected from steamid. The second return value is
//false if they player has no sockets connected
func GetSockets(steamid string) (sockets []*wsevent.Client, success bool) {
	mapMu.RLock()
	defer mapMu.RUnlock()

	sockets, success = steamIDSockets[steamid]
	return
}

//IsConnected returns whether the given steamid is connected to the website
func IsConnected(steamid string) bool {
	_, ok := GetSockets(steamid)
	return ok
}

//ConnectedSockets returns the number of socket connections from steamid
func ConnectedSockets(steamid string) int {
	return len(steamIDSockets[steamid])
}

//Package sessions provides functions to help maintain consistency
//across multiple websocket connections from a single player,
//when the player has multiple tabs/windows open (since each tab opens a new websocket connection)
package sessions

import (
	"sync"
	"time"

	"github.com/TF2Stadium/wsevent"
)

var (
	socketsMu        = new(sync.RWMutex)
	steamIDSockets   = make(map[string][]*wsevent.Client) //steamid -> client array, since players can have multiple tabs open
	socketSpectating = make(map[string]uint)              //socketid -> id of lobby the socket is spectating
	connectedMu      = new(sync.Mutex)
	connectedTimer   = make(map[string](*time.Timer))
)

//AddSocket adds so to the list of sockets connected from steamid
func AddSocket(steamid string, so *wsevent.Client) {
	socketsMu.Lock()
	defer socketsMu.Unlock()

	steamIDSockets[steamid] = append(steamIDSockets[steamid], so)
	if len(steamIDSockets[steamid]) == 1 {
		connectedMu.Lock()
		timer, ok := connectedTimer[steamid]
		if ok {
			timer.Stop()
			delete(connectedTimer, steamid)
		}
		connectedMu.Unlock()
	}
}

//RemoveSocket removes so from the list of sockets connected from steamid
func RemoveSocket(sessionID, steamID string) {
	socketsMu.Lock()
	defer socketsMu.Unlock()

	clients := steamIDSockets[steamID]
	for i, socket := range clients {
		if socket.ID == sessionID {
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
	socketsMu.RLock()
	defer socketsMu.RUnlock()

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
	socketsMu.RLock()
	l := len(steamIDSockets[steamid])
	socketsMu.RUnlock()

	return l
}

//AfterDisconnectedFunc waits the duration to elapse, and if the player with the given
//steamid is still disconnected, calls f in it's own goroutine.
func AfterDisconnectedFunc(steamid string, d time.Duration, f func()) {
	connectedMu.Lock()
	connectedTimer[steamid] = time.AfterFunc(d, func() {
		if !IsConnected(steamid) {
			f()
		}

		connectedMu.Lock()
		delete(connectedTimer, steamid)
		connectedMu.Unlock()
	})
	connectedMu.Unlock()
}

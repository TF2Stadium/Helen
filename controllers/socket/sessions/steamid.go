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
	mapMu = new(sync.RWMutex)
	//steamid -> client array, since players can have multiple tabs open
	steamIDSockets = make(map[string][]*wsevent.Client)
	//socketid -> id of lobby the socket is spectating
	socketSpectating = make(map[string]uint)
	steamIDConnected = make(map[string](chan struct{}))
)

//AddSocket adds so to the list of sockets connected from steamid
func AddSocket(steamid string, so *wsevent.Client) {
	mapMu.Lock()
	defer mapMu.Unlock()

	steamIDSockets[steamid] = append(steamIDSockets[steamid], so)
	if len(steamIDSockets[steamid]) == 1 {
		stop, ok := steamIDConnected[steamid]
		if ok {
			stop <- struct{}{}
		}
	}
}

//RemoveSocket removes so from the list of sockets connected from steamid
func RemoveSocket(sessionID, steamID string) {
	mapMu.Lock()
	defer mapMu.Unlock()

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

//AfterDisconnectedFunc waits the duration to elapse, and if the player with the given
//steamid is still disconnected, calls f in it's own goroutine.
func AfterDisconnectedFunc(steamid string, d time.Duration, f func()) {
	mapMu.Lock()
	stop := make(chan struct{}, 1)
	steamIDConnected[steamid] = stop
	mapMu.Unlock()

	c := time.After(d)

	go func() {
		select {
		case <-c:
			if !IsConnected(steamid) {
				f()
			}
		case <-stop:
			return
		}

		mapMu.Lock()
		delete(steamIDConnected, steamid)
		mapMu.Unlock()
	}()
}

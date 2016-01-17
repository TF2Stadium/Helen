package broadcaster

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

func AddSocket(steamid string, so *wsevent.Client) {
	mapMu.Lock()
	defer mapMu.Unlock()

	steamIDSockets[steamid] = append(steamIDSockets[steamid], so)
}

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

func GetSockets(steamid string) (sockets []*wsevent.Client, success bool) {
	mu.RLock()
	defer mu.RUnlock()

	sockets, success = steamIDSockets[steamid]
	return
}

func IsConnected(steamid string) bool {
	_, ok := GetSockets(steamid)
	return ok
}

func ConnectedSockets(steamid string) int {
	return len(steamIDSockets[steamid])
}

func SetSpectator(socketID string, lobbyID uint) {
	mapMu.Lock()
	defer mapMu.Unlock()
	socketSpectating[socketID] = lobbyID
}

func GetSpectating(socketID string) (lobbyID uint, ok bool) {
	mapMu.RLock()
	defer mapMu.RUnlock()
	lobbyID, ok = socketSpectating[socketID]
	return
}

func RemoveSpectator(socketID string) {
	mapMu.Lock()
	defer mapMu.Unlock()
	delete(socketSpectating, socketID)
}

func IsSpectating(socketID string, lobbyID uint) bool {
	mapMu.RLock()
	defer mapMu.RUnlock()
	_, ok := socketSpectating[socketID]
	return ok
}

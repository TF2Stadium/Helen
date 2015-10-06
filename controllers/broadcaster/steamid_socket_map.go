package broadcaster

import (
	"github.com/googollee/go-socket.io"
	"sync"
)

var steamIdSocketMap = make(map[string]socketio.Socket)
var steamIdSocketMapLock sync.Mutex

func SetSocket(steamid string, so socketio.Socket) {
	steamIdSocketMapLock.Lock()
	defer steamIdSocketMapLock.Unlock()

	steamIdSocketMap[steamid] = so
}

func RemoveSocket(steamid string) {
	steamIdSocketMapLock.Lock()
	defer steamIdSocketMapLock.Unlock()

	delete(steamIdSocketMap, steamid)
}

func GetSocket(steamid string) (so socketio.Socket, success bool) {
	steamIdSocketMapLock.Lock()
	defer steamIdSocketMapLock.Unlock()

	so, success = steamIdSocketMap[steamid]
	return
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package broadcaster

import (
	"sync"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/wsevent"
)

var mu = new(sync.RWMutex)
var socketServer *wsevent.Server
var socketServerNoLogin *wsevent.Server

func Init(server *wsevent.Server, nologin *wsevent.Server) {
	mu.Lock()
	socketServer = server
	socketServerNoLogin = nologin
	mu.Unlock()
}

func SendMessage(steamid string, event string, content interface{}) {
	socket, ok := GetSocket(steamid)
	if !ok {
		return
	}

	mu.RLock()
	defer mu.RUnlock()
	socket.EmitJSON(helpers.NewRequest(event, content))
}

func SendMessageToRoom(room string, event string, content interface{}) {
	if socketServer == nil {
		return
	}

	v := helpers.NewRequest(event, content)

	mu.RLock()
	defer mu.RUnlock()
	socketServer.BroadcastJSON(room, v)
	socketServerNoLogin.BroadcastJSON(room, v)
}

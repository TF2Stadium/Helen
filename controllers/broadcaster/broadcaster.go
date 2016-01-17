// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package broadcaster

import (
	"sync"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/wsevent"
)

var (
	mu                  = new(sync.RWMutex) //protects the following vars
	socketServer        *wsevent.Server
	socketServerNoLogin *wsevent.Server
)

func Init(server *wsevent.Server, nologin *wsevent.Server) {
	mu.Lock()
	socketServer = server
	socketServerNoLogin = nologin
	mu.Unlock()
}

func SendMessage(steamid string, event string, content interface{}) {
	sockets, ok := GetSockets(steamid)
	if !ok {
		return
	}

	mu.RLock()
	defer mu.RUnlock()
	for _, socket := range sockets {
		go func(so *wsevent.Client) {
			so.EmitJSON(helpers.NewRequest(event, content))
		}(socket)
	}
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

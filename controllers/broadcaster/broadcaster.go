// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package broadcaster

import (
	"sync"

	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/wsevent"
)

var (
	mu           = new(sync.RWMutex) //protects the following vars
	auth, noauth *wsevent.Server
)

func Init(a, na *wsevent.Server) {
	mu.Lock()
	auth = a
	noauth = a
	mu.Unlock()
}

func SendMessage(steamid string, event string, content interface{}) {
	sockets, ok := sessions.GetSockets(steamid)
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
	if auth == nil { //check for testing purposes
		return
	}

	v := helpers.NewRequest(event, content)

	mu.RLock()
	defer mu.RUnlock()
	auth.BroadcastJSON(room, v)
	noauth.BroadcastJSON(room, v)
}

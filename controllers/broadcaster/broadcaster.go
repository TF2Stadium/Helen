// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package broadcaster

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/wsevent"
)

var socketServer *wsevent.Server
var socketServerNoLogin *wsevent.Server

func Init(server *wsevent.Server, nologin *wsevent.Server) {
	socketServer = server
	socketServerNoLogin = nologin
}

func SendMessage(steamid string, event string, content string) {
	socket, ok := GetSocket(steamid)
	if !ok {
		return
	}
	socket.EmitJSON(helpers.NewRequest(event, content))
}

func SendMessageToRoom(room string, event string, content string) {
	if socketServer == nil {
		return
	}

	v := helpers.NewRequest(event, content)

	socketServer.BroadcastJSON(room, v)
	socketServerNoLogin.BroadcastJSON(room, v)
}

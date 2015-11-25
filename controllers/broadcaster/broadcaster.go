// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package broadcaster

import (
	"github.com/TF2Stadium/Helen/helpers"
)

type commonBroadcaster interface {
	BroadcastJSON(string, interface{})
}

var socketServer commonBroadcaster
var socketServerNoLogin commonBroadcaster

func Init(server commonBroadcaster, nologin commonBroadcaster) {
	socketServer = server
	socketServerNoLogin = nologin
}

func SendMessage(steamid string, event string, content string) {
	socket, ok := GetSocket(steamid)
	if !ok {
		helpers.Logger.Critical("Failed to get the user's socket: %s", steamid)
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

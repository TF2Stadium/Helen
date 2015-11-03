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

func Init(server commonBroadcaster) {
	socketServer = server
}

func SendMessage(steamid string, event string, content string) {
	socket, ok := GetSocket(steamid)
	if !ok {
		helpers.Logger.Critical("Failed to get the user's socket: %s", steamid)
		return
	}
	socket.EmitJSON(helpers.NewRequest(event, []byte(content)))
}

func SendMessageToRoom(room string, event string, content string) {
	socketServer.BroadcastJSON(room, helpers.NewRequest(event, []byte(content)))
}

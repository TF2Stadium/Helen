// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package broadcaster

import (
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

func SendMessage(steamid string, event string, content interface{}) {
	sockets, ok := sessions.GetSockets(steamid)
	if !ok {
		return
	}

	for _, socket := range sockets {
		go func(so *wsevent.Client) {
			so.EmitJSON(helpers.NewRequest(event, content))
		}(socket)
	}

	return
}

func SendMessageToRoom(r string, event string, content interface{}) {
	v := helpers.NewRequest(event, content)

	socket.AuthServer.BroadcastJSON(r, v)
	socket.UnauthServer.BroadcastJSON(r, v)
}

func SendMessageSkipIDs(skipID, steamid, event string, content interface{}) {
	sockets, ok := sessions.GetSockets(steamid)
	if !ok {
		return
	}

	for _, socket := range sockets {
		if socket.ID != skipID {
			socket.EmitJSON(helpers.NewRequest(event, content))
		}
	}
}

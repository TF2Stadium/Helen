// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/googollee/go-socket.io"
)

func SocketMockUpInit(so socketio.Socket) {

	helpers.Logger.Debug("on connection")

	so.Join("chat")

	so.On("chat message", func(msg string) {
		helpers.Logger.Debug("emit:", so.Emit("chat message", msg))
		so.BroadcastTo("chat", "chat message", msg)
	})

	so.On("disconnection", func() {
		helpers.Logger.Debug("on disconnect")
	})
}

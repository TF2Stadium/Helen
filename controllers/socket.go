package controllers

import (
	"log"

	"github.com/googollee/go-socket.io"
)

func SocketInit(so socketio.Socket) {
	log.Println("on connection")
	so.Join("chat")
	so.On("chat message", func(msg string) {
		log.Println("emit:", so.Emit("chat message", msg))
		so.BroadcastTo("chat", "chat message", msg)
	})
	so.On("disconnection", func() {
		log.Println("on disconnect")
	})
}

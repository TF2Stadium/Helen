package controllerhelpers

import (
	"container/ring"
	"sync"

	"github.com/googollee/go-socket.io"
)

type chatRing struct {
	curr  *ring.Ring
	first *ring.Ring
	*sync.Mutex
}

var chatScrollback *chatRing

func InitChatScrollback() {
	r := ring.New(20)
	chatScrollback = &chatRing{r, r, new(sync.Mutex)}
}

func AddScrollbackMessage(message string) {
	chatScrollback.Lock()
	defer chatScrollback.Unlock()

	if chatScrollback.curr.Value != nil {
		chatScrollback.first = chatScrollback.first.Next()
	}

	chatScrollback.curr.Value = message
	chatScrollback.curr = chatScrollback.curr.Next()
}

func BroadcastScrollback(so socketio.Socket) {
	chatScrollback.Lock()
	defer chatScrollback.Unlock()

	curr := chatScrollback.first
	if curr.Value == nil {
		return
	}

	for printed := 0; printed != 20; printed++ {
		if curr.Value == nil {
			return
		}
		so.Emit("chatReceive", curr.Value.(string))
		curr = curr.Next()
	}
}

package controllerhelpers

import (
	"container/ring"
	"sync"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/vibhavp/wsevent"
)

type chatRing struct {
	curr  *ring.Ring
	first *ring.Ring
	*sync.Mutex
}

var chatScrollback = make(map[uint]*chatRing)

func initChatScrollback(room uint) *chatRing {
	r := ring.New(20)
	return &chatRing{r, r, new(sync.Mutex)}
}

func AddScrollbackMessage(room uint, message string) {
	if _, ok := chatScrollback[room]; !ok {
		chatScrollback[room] = initChatScrollback(room)
	}
	c := chatScrollback[room]

	c.Lock()
	defer c.Unlock()

	if c.curr.Value != nil {
		c.first = c.first.Next()
	}

	c.curr.Value = message
	c.curr = c.curr.Next()
}

func BroadcastScrollback(so *wsevent.Client, room uint) {

	so.EmitJSON(helpers.NewRequest("chatHistoryClear", []byte("{}")))

	c, ok := chatScrollback[room]
	if !ok {
		return
	}

	c.Lock()
	defer c.Unlock()

	curr := c.first
	if curr.Value == nil {
		return
	}

	for printed := 0; printed != 20; printed++ {
		if curr.Value == nil {
			return
		}
		so.EmitJSON(helpers.NewRequest("chatReceive", []byte(curr.Value.(string))))
		curr = curr.Next()
	}
}

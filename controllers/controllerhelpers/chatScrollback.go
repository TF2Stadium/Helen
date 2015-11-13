package controllerhelpers

import (
	"container/ring"
	"sync"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/wsevent"
)

type chatRing struct {
	curr  *ring.Ring
	first *ring.Ring
	*sync.RWMutex
}

var mapLock = new(sync.RWMutex)
var chatScrollback = make(map[uint]*chatRing)

func initChatScrollback() *chatRing {
	r := ring.New(20)
	return &chatRing{r, r, new(sync.RWMutex)}
}

func AddScrollbackMessage(room uint, message string) {
	mapLock.Lock()
	if _, ok := chatScrollback[room]; !ok {
		chatScrollback[room] = initChatScrollback()
	}
	c := chatScrollback[room]
	mapLock.Unlock()

	c.Lock()
	defer c.Unlock()

	if c.curr.Value != nil {
		c.first = c.first.Next()
	}

	c.curr.Value = message
	c.curr = c.curr.Next()
}

func BroadcastScrollback(so *wsevent.Client, room uint) {

	so.EmitJSON(helpers.NewRequest("chatHistoryClear", "{}"))

	mapLock.RLock()
	c, ok := chatScrollback[room]
	mapLock.RUnlock()
	if !ok {
		return
	}

	c.RLock()
	defer c.RUnlock()

	curr := c.first
	if curr.Value == nil {
		return
	}

	for printed := 0; printed != 20; printed++ {
		if curr.Value == nil {
			return
		}
		so.EmitJSON(helpers.NewRequest("chatReceive", curr.Value.(string)))
		curr = curr.Next()
	}
}

package controllerhelpers

import (
	"encoding/json"
	"sync"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/wsevent"
)

type chatRing struct {
	messages []string
	curr     int32
	*sync.RWMutex
}

type ChatHistoryClearEvent struct {
	Room uint `json:"room"`
}

var mapLock = new(sync.RWMutex)
var chatScrollback = make(map[uint]*chatRing)

func initChatScrollback() *chatRing {
	return &chatRing{make([]string, 20), 0, new(sync.RWMutex)}
}

func AddScrollbackMessage(room uint, message string) {
	mapLock.Lock()
	if _, ok := chatScrollback[room]; !ok {
		chatScrollback[room] = initChatScrollback()
	}
	c := chatScrollback[room]
	mapLock.Unlock()

	c.Lock()
	c.messages[c.curr] = message
	c.curr += 1
	if c.curr == 20 {
		c.curr = 0
	}
	c.Unlock()
}

func BroadcastScrollback(so *wsevent.Client, room uint) {
	bytes, _ := json.Marshal(ChatHistoryClearEvent{room})
	so.EmitJSON(helpers.NewRequest("chatHistoryClear", string(bytes)))

	mapLock.RLock()
	c, ok := chatScrollback[room]
	mapLock.RUnlock()
	if !ok {
		return
	}

	c.RLock()
	defer c.RUnlock()

	curr := c.curr
	if c.messages[curr] == "" {
		curr = 0
	}

	for i := 0; i < 20; i++ {
		if c.messages[curr] == "" {
			return
		}

		so.EmitJSON(helpers.NewRequest("chatReceive", c.messages[curr]))
		curr += 1
		if curr == 20 {
			curr = 0
		}
	}
}

package controllerhelpers

import (
	"encoding/json"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

func BroadcastScrollback(so *wsevent.Client, room uint) {
	// bytes, _ := json.Marshal(ChatHistoryClearEvent{room})
	// so.EmitJSON(helpers.NewRequest("chatHistoryClear", string(bytes)))

	messages, err := models.GetScrollback(int(room))
	if err != nil {
		return
	}

	for i := len(messages) - 1; i != -1; i-- {
		bytes, _ := json.Marshal(messages[i])

		so.EmitJSON(helpers.NewRequest("chatReceive", string(bytes)))
	}
}

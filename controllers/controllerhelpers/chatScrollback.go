package controllerhelpers

import (
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

	so.EmitJSON(helpers.NewRequest("chatScrollback", messages))
}

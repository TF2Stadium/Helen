package controllerhelpers

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/chat"
	"github.com/TF2Stadium/wsevent"
)

func BroadcastScrollback(so *wsevent.Client, room uint) {
	messages, err := chat.GetScrollback(int(room))
	if err != nil {
		return
	}

	so.EmitJSON(helpers.NewRequest("chatScrollback", messages))
}

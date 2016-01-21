package socket

import (
	"encoding/json"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/routes/socket/middleware"
	"github.com/TF2Stadium/wsevent"
)

var (
	//AuthServer is the wsevent server where authenticated (logged in) users/sockets
	//are added to
	AuthServer = wsevent.NewServer(middleware.JSONCodec{})
	//UnauthServer is the wsevent server where unauthenticated users/sockets
	//are added to
	UnauthServer = wsevent.NewServer(middleware.JSONCodec{})
)

func InitializeSocketServer() {
	AuthServer.DefaultHandler = func(_ *wsevent.Client, _ []byte) interface{} {
		return helpers.NewTPError("No such request.", -3)
	}

	UnauthServer.DefaultHandler = func(_ *wsevent.Client, data []byte) interface{} {
		return helpers.NewTPError("Player isn't logged in.", -4)
	}

}

func getEvent(data []byte) string {
	var js struct {
		Request string
	}
	json.Unmarshal(data, &js)
	return js.Request
}

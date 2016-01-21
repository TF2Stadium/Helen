package socket

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/routes/socket/middleware"
	"github.com/TF2Stadium/wsevent"
)

var (
	//AuthServer is the wsevent server where authenticated (logged in) users/sockets
	//are added to
	AuthServer = wsevent.NewServer(middleware.JSONCodec{}, func(_ *wsevent.Client, _ struct{}) interface{} {
		return helpers.NewTPError("No such request.", -3)
	})
	//UnauthServer is the wsevent server where unauthenticated users/sockets
	//are added to
	UnauthServer = wsevent.NewServer(middleware.JSONCodec{}, func(_ *wsevent.Client, _ struct{}) interface{} {
		return helpers.NewTPError("Player isn't logged in.", -4)
	})
)

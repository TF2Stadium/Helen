package socket

import (
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/socket/handler"
	"github.com/TF2Stadium/Helen/internal/pprof"
	"github.com/TF2Stadium/Helen/routes/socket"
)

func RegisterHandlers() {
	socket.AuthServer.OnDisconnect = hooks.OnDisconnect
	socket.UnauthServer.OnDisconnect = func(string) { pprof.Clients.Add(-1) }

	socket.AuthServer.Register(handler.Global{}) //Global Handlers
	socket.AuthServer.Register(handler.Lobby{})  //Lobby Handlers
	socket.AuthServer.Register(handler.Player{}) //Player Handlers
	socket.AuthServer.Register(handler.Chat{})   //Chat Handlers

	socket.UnauthServer.Register(handler.Unauth{})
}

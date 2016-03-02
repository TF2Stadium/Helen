package socket

import (
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/socket/handler"
	"github.com/TF2Stadium/Helen/internal/pprof"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/dgrijalva/jwt-go"
)

func RegisterHandlers() {
	socket.AuthServer.OnDisconnect = hooks.OnDisconnect
	socket.UnauthServer.OnDisconnect = func(string, *jwt.Token) { pprof.Clients.Add(-1) }

	socket.AuthServer.Register(handler.Global{}) //Global Handlers
	socket.AuthServer.Register(handler.Lobby{})  //Lobby Handlers
	socket.AuthServer.Register(handler.Player{}) //Player Handlers
	socket.AuthServer.Register(handler.Chat{})   //Chat Handlers
	socket.AuthServer.Register(handler.Serveme{})
	socket.AuthServer.Register(handler.Mumble{})

	socket.UnauthServer.Register(handler.Unauth{})
}

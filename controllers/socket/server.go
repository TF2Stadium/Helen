package socket

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/socket/handler"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

func RegisterHandlers() {
	socket.AuthServer.OnDisconnect = hooks.OnDisconnect
	socket.UnauthServer.OnDisconnect = hooks.OnDisconnect

	socket.AuthServer.Register(handler.Global{}) //Global Handlers
	socket.AuthServer.Register(handler.Lobby{})  //Lobby Handlers
	socket.AuthServer.Register(handler.Player{}) //Player Handlers
	socket.AuthServer.Register(handler.Chat{})   //Chat Handlers

	socket.UnauthServer.On("lobbySpectatorJoin", unAuthSpecJoin)
}

func unAuthSpecJoin(so *wsevent.Client, args struct {
	ID *uint `json:"id"`
}) interface{} {

	var lob *models.Lobby
	lob, tperr := models.GetLobbyByID(*args.ID)

	if tperr != nil {
		return tperr
	}

	hooks.AfterLobbySpec(socket.UnauthServer, so, lob)

	so.EmitJSON(helpers.NewRequest("lobbyData", models.DecorateLobbyData(lob, true)))

	return chelpers.EmptySuccessJS

}

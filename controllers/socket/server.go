package socket

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/socket/internal/handler"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

var (
	//AuthServer is the wsevent server where authenticated (logged in) users/sockets
	//are added to
	AuthServer = wsevent.NewServer()
	//UnauthServer is the wsevent server where unauthenticated users/sockets
	//are added to
	UnauthServer = wsevent.NewServer()
)

func init() {
	serverInit(AuthServer, UnauthServer)
}

//NewServers replaces AuthServer and UnauthServer with empty wsevent servers.
//Use it ONLY for testing purposes
func NewServers() {
	AuthServer = wsevent.NewServer()
	UnauthServer = wsevent.NewServer()
	serverInit(AuthServer, UnauthServer)
}

func serverInit(server *wsevent.Server, noAuthServer *wsevent.Server) {
	server.OnDisconnect = hooks.OnDisconnect
	server.Extractor = getEvent

	noAuthServer.OnDisconnect = hooks.OnDisconnect
	noAuthServer.Extractor = getEvent

	server.Register(handler.Global{}) //Global Handlers
	server.Register(handler.Lobby{})  //Lobby Handlers
	server.Register(handler.Player{}) //Player Handlers
	server.Register(handler.Chat{})   //Chat Handlers

	server.DefaultHandler = func(_ *wsevent.Server, _ *wsevent.Client, _ []byte) interface{} {
		return helpers.NewTPError("No such request.", -3)
	}

	noAuthServer.On("lobbySpectatorJoin", unAuthSpecJoin)
	noAuthServer.On("getSocketInfo", (handler.Global{}).GetSocketInfo)

	noAuthServer.DefaultHandler = func(_ *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
		return helpers.NewTPError("Player isn't logged in.", -4)
	}
}

func unAuthSpecJoin(s *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	var args struct {
		ID *uint `json:"id"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	var lob *models.Lobby
	lob, tperr := models.GetLobbyByID(*args.ID)

	if tperr != nil {
		return tperr
	}

	hooks.AfterLobbySpec(s, so, lob)

	so.EmitJSON(helpers.NewRequest("lobbyData", models.DecorateLobbyData(lob, true)))

	return chelpers.EmptySuccessJS

}

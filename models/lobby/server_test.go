package lobby

import (
	"testing"

	"github.com/TeamPlayTF/Server/config"
	"github.com/TeamPlayTF/Server/models/lobby"
)

// change this if you wanna test the server
// make sure you have the server running at the moment
var shouldTest bool = true
var svr *lobby.Server

func TestServerSetup(t *testing.T) {
	if shouldTest {
		config.SetupConstants()
		lobby.InitConfigs()

		svr = lobby.NewServer()
		svr.Map = "cp_process_final"
		svr.Type = lobby.LobbyTypeHighlander
		svr.League = "ugc"
		svr.LobbyId = 0
		svr.Address = "192.168.1.94:27015"
		svr.RconPassword = "30035"
		svr.LobbyPassword = "12345"
		svr.AllowPlayer("[U:1:106866319]") // your steamId, so it wont be kicking you out everytime
		svr.Setup()
	}
}

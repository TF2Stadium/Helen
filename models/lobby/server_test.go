package lobby

import (
	"errors"
	"log"
	"testing"

	"github.com/TeamPlayTF/Server/config"
	"github.com/TeamPlayTF/Server/models/lobby"
)

// change this if you wanna test the server
// make sure you have the server running at the moment
var shouldTest bool = false
var svr *lobby.Server

func TestServerSetup(t *testing.T) {
	if shouldTest {
		config.SetupConstants()
		lobby.InitConfigs()

		commId := "76561198067132047" // your commId, so it wont be kicking you out everytime

		svr = lobby.NewServer()
		svr.Map = "cp_process_final"
		svr.Type = lobby.LobbyTypeHighlander
		svr.League = lobby.LeagueUgc
		svr.Address = "192.168.1.94:27015"
		svr.RconPassword = "rconPassword"
		svr.LobbyPassword = "12345"

		svr.AllowPlayer(commId)

		setupErr := svr.Setup()

		if setupErr != nil {
			log.Fatal(setupErr)
		}

		playerIsAllowed := svr.IsPlayerAllowed(commId)

		if playerIsAllowed == false {
			log.Fatal(errors.New("Error: your player isn't allowed in the server"))
		}

		svr.End()
	}
}

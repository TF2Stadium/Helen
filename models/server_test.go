package models

import (
	"testing"

	"github.com/TF2Stadium/Server/config"
	"github.com/stretchr/testify/assert"
)

// change this if you wanna test the server
// make sure you have the server running at the moment
var shouldTest bool = false
var svr *Server

func TestServerSetup(t *testing.T) {
	if shouldTest {
		config.SetupConstants()
		InitServerConfigs()

		commId := "76561198067132047" // your commId, so it wont be kicking you out everytime

		svr = NewServer()
		svr.Map = "cp_process_final"
		svr.Type = LobbyTypeHighlander
		svr.League = LeagueUgc
		svr.Address = "192.168.1.94:27015"
		svr.RconPassword = "rconPassword"
		svr.LobbyPassword = "12345"

		svr.AllowPlayer(commId)

		setupErr := svr.Setup()
		assert.Nil(t, setupErr, "Setup error")

		playerIsAllowed := svr.IsPlayerAllowed(commId)
		assert.True(t, playerIsAllowed, "Player isn't allowed, he should")

		svr.DisallowPlayer(commId)

		playerIsntAllowed := svr.IsPlayerAllowed(commId)
		assert.False(t, playerIsntAllowed, "Player is allowed, he shouldn't")

		svr.End()
	}
}

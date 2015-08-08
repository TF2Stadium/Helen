package models

import (
	"os"
	"testing"

	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/stretchr/testify/assert"
)

// change this if you wanna test the server
// make sure you have the server running at the moment
var host = os.Getenv("TEST_TF2_SERVER_HOST")
var password = os.Getenv("TEST_TF2_SERVER_PASSWORD")
var shouldTest bool = host != "" && password != ""

var shouldTestLive = false

var svr *Server

func init() {
	helpers.InitLogger()
}

func TestServerSetup(t *testing.T) {
	if shouldTestLive {
		config.SetupConstants()
		InitServerConfigs()

		commId := "76561198067132047" // your commId, so it wont be kicking you out everytime

		info := ServerRecord{
			Host:         host,
			RconPassword: password,
		}

		svr = NewServer()
		svr.Map = "cp_process_final"
		svr.Type = LobbyTypeHighlander
		svr.League = LeagueUgc
		svr.ServerPassword = "12345"
		svr.Info = info

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

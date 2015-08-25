package models

import (
	"errors"
	"net/rpc"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
)

type ServerBootstrap struct {
	LobbyId       uint
	Info          ServerRecord
	Players       []string
	BannedPlayers []string
}

type Event map[string]interface{}

type Args struct {
	Id      uint
	Info    ServerRecord
	Type    LobbyType
	League  string
	Map     string
	SteamId string
}

var EventQueueEmptyError = errors.New("Event queue empty")
var Pauling *rpc.Client

func PaulingConnect() {
	helpers.Logger.Debug("Connecting to Pauling on port %s", config.Constants.PaulingPort)
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	Pauling = client
	helpers.Logger.Debug("Connected!")
}

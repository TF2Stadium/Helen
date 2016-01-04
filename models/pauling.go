// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"net/rpc"
	"sync"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
)

type ServerBootstrap struct {
	LobbyId       uint
	Info          ServerRecord
	Players       []string
	BannedPlayers []string
}

type Args struct {
	Id        uint
	Info      ServerRecord
	Type      LobbyType
	League    string
	Whitelist string
	Map       string
	SteamId   string
	SteamId2  string
	Slot      string
	Text      string
}

var mu = new(sync.RWMutex)
var pauling *rpc.Client

func paulingReconnect() {
	if config.Constants.ServerMockUp {
		return
	}

	mu.Lock()
	defer mu.Unlock()
	helpers.Logger.Debug("Reconnecting to Pauling on port %s", config.Constants.PaulingPort)
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	for err != nil {
		helpers.Logger.Critical("%s", err.Error())
		time.Sleep(1 * time.Second)
		client, err = rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	}

	pauling = client
	helpers.Logger.Debug("Connected!")
}

func PaulingConnect() {
	if config.Constants.ServerMockUp {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var err error
	helpers.Logger.Debug("Connecting to Pauling on port %s", config.Constants.PaulingPort)
	pauling, err = rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	helpers.Logger.Debug("Connected!")
	pauling.Call("Pauling.Connect", config.Constants.RPCPort, struct{}{})
}

func DisallowPlayer(lobbyId uint, steamId string) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	mu.RLock()
	defer mu.RUnlock()

	return pauling.Call("Pauling.DisallowPlayer", &Args{Id: lobbyId, SteamId: steamId}, &Args{})
}

func SetupServer(lobbyId uint, info ServerRecord, lobbyType LobbyType, league string,
	whitelist string, mapName string) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	mu.RLock()
	defer mu.RUnlock()

	args := &Args{
		Id:        lobbyId,
		Info:      info,
		Type:      lobbyType,
		League:    league,
		Whitelist: whitelist,
		Map:       mapName}
	return pauling.Call("Pauling.SetupServer", args, &Args{})
}

func ReExecConfig(lobbyId uint) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	return pauling.Call("Pauling.ReExecConfig", &Args{Id: lobbyId}, &Args{})
}

func VerifyInfo(info ServerRecord) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	mu.RLock()
	defer mu.RUnlock()

	return pauling.Call("Pauling.VerifyInfo", &info, &Args{})
}

func IsPlayerInServer(steamid string) (reply bool) {
	if config.Constants.ServerMockUp {
		return false
	}

	mu.RLock()
	defer mu.RUnlock()

	args := &Args{SteamId: steamid}
	pauling.Call("Pauling.IsPlayerInServer", &args, &reply)

	return
}

func End(lobbyId uint) {
	if config.Constants.ServerMockUp {
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	pauling.Call("Pauling.End", &Args{Id: lobbyId}, &Args{})
}

func Say(lobbyId uint, text string) {
	if config.Constants.ServerMockUp {
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	pauling.Call("Pauling.Say", &Args{Id: lobbyId, Text: text}, &Args{})
}

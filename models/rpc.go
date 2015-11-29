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
	Whitelist int
	Map       string
	SteamId   string
	SteamId2  string
	Slot      string
}

var PaulingLock = new(sync.RWMutex)
var Pauling *rpc.Client

type Event map[string]interface{}

func PaulingReconnect() {
	if config.Constants.ServerMockUp {
		return
	}

	PaulingLock.Lock()
	defer PaulingLock.Unlock()
	helpers.Logger.Debug("Reconnecting to Pauling on port %s", config.Constants.PaulingPort)
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	for err != nil {
		helpers.Logger.Critical("%s", err.Error())
		time.Sleep(1 * time.Second)
		client, err = rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	}

	Pauling = client
	helpers.Logger.Debug("Connected!")
}

func PaulingConnect() {
	if config.Constants.ServerMockUp {
		return
	}

	PaulingLock.Lock()
	defer PaulingLock.Unlock()

	helpers.Logger.Debug("Connecting to Pauling on port %s", config.Constants.PaulingPort)
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	Pauling = client
	helpers.Logger.Debug("Connected!")
}

func AllowPlayer(lobbyId uint, steamId string, slot string) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	PaulingLock.RLock()
	defer PaulingLock.RUnlock()

	return Pauling.Call("Pauling.AllowPlayer", &Args{Id: lobbyId, SteamId: steamId, Slot: slot}, &Args{})
}

func DisallowPlayer(lobbyId uint, steamId string) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	PaulingLock.RLock()
	defer PaulingLock.RUnlock()

	return Pauling.Call("Pauling.DisallowPlayer", &Args{Id: lobbyId, SteamId: steamId}, &Args{})
}

func SetupServer(lobbyId uint, info ServerRecord, lobbyType LobbyType, league string,
	whitelist int, mapName string) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	PaulingLock.RLock()
	defer PaulingLock.RUnlock()

	args := &Args{
		Id:        lobbyId,
		Info:      info,
		Type:      lobbyType,
		League:    league,
		Whitelist: whitelist,
		Map:       mapName}
	return Pauling.Call("Pauling.SetupServer", args, &Args{})
}

func ReExecConfig(lobbyId uint) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	return Pauling.Call("Pauling.ReExecConfig", &Args{Id: lobbyId}, &Args{})
}

func VerifyInfo(info ServerRecord) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	PaulingLock.RLock()
	defer PaulingLock.RUnlock()

	return Pauling.Call("Pauling.VerifyInfo", &info, &Args{})
}

func IsPlayerInServer(steamid string) (reply bool) {
	if config.Constants.ServerMockUp {
		return false
	}

	PaulingLock.RLock()
	defer PaulingLock.RUnlock()

	args := &Args{SteamId: steamid}
	Pauling.Call("Pauling.IsPlayerInServer", &args, &reply)

	return
}

func End(lobbyId uint) {
	if config.Constants.ServerMockUp {
		return
	}

	PaulingLock.RLock()
	defer PaulingLock.RUnlock()

	Pauling.Call("Pauling.End", &Args{Id: lobbyId}, &Args{})
}

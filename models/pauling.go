// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"net/rpc"
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

func call(method string, args, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	if err != nil {
		for err != nil {
			time.Sleep(1 * time.Second)
			client, err = rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
		}
	}

	err = client.Call(method, args, reply)
	client.Close()
	return err
}

func CheckConnection() {
	err := call("Pauling.Test", struct{}{}, &struct{}{})
	if err != nil {
		helpers.Logger.Fatal(err.Error())
	}

	helpers.Logger.Debug("Able to connect to Pauling")
}

func DisallowPlayer(lobbyId uint, steamId string) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	return call("Pauling.DisallowPlayer", &Args{Id: lobbyId, SteamId: steamId}, &Args{})
}

func SetupServer(lobbyId uint, info ServerRecord, lobbyType LobbyType, league string,
	whitelist string, mapName string) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	args := &Args{
		Id:        lobbyId,
		Info:      info,
		Type:      lobbyType,
		League:    league,
		Whitelist: whitelist,
		Map:       mapName}
	return call("Pauling.SetupServer", args, &Args{})
}

func ReExecConfig(lobbyId uint) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	return call("Pauling.ReExecConfig", &Args{Id: lobbyId}, &Args{})
}

func VerifyInfo(info ServerRecord) error {
	if config.Constants.ServerMockUp {
		return nil
	}

	return call("Pauling.VerifyInfo", &info, &Args{})
}

func IsPlayerInServer(steamid string) (reply bool) {
	if config.Constants.ServerMockUp {
		return false
	}

	args := &Args{SteamId: steamid}
	call("Pauling.IsPlayerInServer", &args, &reply)

	return
}

func End(lobbyId uint) {
	if config.Constants.ServerMockUp {
		return
	}

	call("Pauling.End", &Args{Id: lobbyId}, &Args{})
}

func Say(lobbyId uint, text string) {
	if config.Constants.ServerMockUp {
		return
	}

	call("Pauling.Say", &Args{Id: lobbyId, Text: text}, &Args{})
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"time"

	"github.com/TF2Stadium/Helen/config"
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

func DisallowPlayer(lobbyId uint, steamId string) error {
	return call(config.Constants.PaulingAddr, "Pauling.DisallowPlayer", &Args{Id: lobbyId, SteamId: steamId}, &Args{})
}

func SetupServer(lobbyId uint, info ServerRecord, lobbyType LobbyType, league string,
	whitelist string, mapName string) error {

	args := &Args{
		Id:        lobbyId,
		Info:      info,
		Type:      lobbyType,
		League:    league,
		Whitelist: whitelist,
		Map:       mapName}
	return call(config.Constants.PaulingAddr, "Pauling.SetupServer", args, &Args{})
}

func ReExecConfig(lobbyId uint) error {
	return call(config.Constants.PaulingAddr, "Pauling.ReExecConfig", &Args{Id: lobbyId}, &Args{})
}

func VerifyInfo(info ServerRecord) error {
	return call(config.Constants.PaulingAddr, "Pauling.VerifyInfo", &info, &Args{})
}

func IsPlayerInServer(steamid string) (reply bool) {
	args := &Args{SteamId: steamid}
	call(config.Constants.PaulingAddr, "Pauling.IsPlayerInServer", &args, &reply)

	return
}

func End(lobbyId uint) {
	call(config.Constants.PaulingAddr, "Pauling.End", &Args{Id: lobbyId}, &Args{})
}

func Say(lobbyId uint, text string) {
	call(config.Constants.PaulingAddr, "Pauling.Say", &Args{Id: lobbyId, Text: text}, &Args{})
}

func serverExists(lobbyID uint) (exists bool) {
	call(config.Constants.PaulingAddr, "Pauling.Exists", lobbyID, &exists)
	return
}

func Ping() {
	tick := time.NewTicker(time.Second)
	for {
		<-tick.C
		call(config.Constants.PaulingAddr, "Pauling.Ping", struct{}{}, &struct{}{})
	}
}

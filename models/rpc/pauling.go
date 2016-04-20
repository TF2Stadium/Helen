// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package rpc

import (
	"github.com/TF2Stadium/Helen/models/gameserver"
	"github.com/TF2Stadium/Helen/models/lobby/format"
)

type Args struct {
	Id        uint
	Info      gameserver.ServerRecord
	Type      format.Format
	League    string
	Whitelist string
	Map       string
	SteamId   string
	SteamId2  string
	Slot      string
	Text      string
	ChangeMap bool
}

func DisallowPlayer(lobbyId uint, steamId string, playerID uint) error {
	if !*paulingDisabled {
		pauling.Call("Pauling.DisallowPlayer", &Args{Id: lobbyId, SteamId: steamId}, &struct{}{})
	}

	if !*fumbleDisabled {
		fumble.Call("Fumble.RemovePlayer", playerID, &struct{}{})
	}

	return nil
}

func SetupServer(lobbyId uint, info gameserver.ServerRecord, lobbyType format.Format, league string,
	whitelist string, mapName string) error {
	if *paulingDisabled {
		return nil
	}

	args := &Args{
		Id:        lobbyId,
		Info:      info,
		Type:      lobbyType,
		League:    league,
		Whitelist: whitelist,
		Map:       mapName}
	return pauling.Call("Pauling.SetupServer", args, &struct{}{})
}

func ReExecConfig(lobbyId uint, changeMap bool) error {
	if *paulingDisabled {
		return nil
	}
	return pauling.Call("Pauling.ReExecConfig", &Args{Id: lobbyId, ChangeMap: changeMap}, &struct{}{})
}

func VerifyInfo(info gameserver.ServerRecord) error {
	if *paulingDisabled {
		return nil
	}
	return pauling.Call("Pauling.VerifyInfo", &info, &struct{}{})
}

func End(lobbyId uint) {
	if *paulingDisabled {
		return
	}
	pauling.Call("Pauling.End", &Args{Id: lobbyId}, &struct{}{})
}

func Say(lobbyId uint, text string) {
	if *paulingDisabled {
		return
	}
	pauling.Call("Pauling.Say", &Args{Id: lobbyId, Text: text}, &struct{}{})
}

func serverExists(lobbyID uint) (exists bool) {
	if *paulingDisabled {
		return false
	}
	pauling.Call("Pauling.Exists", lobbyID, &exists)
	return
}

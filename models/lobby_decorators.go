// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"strconv"

	"github.com/TF2Stadium/Helen/config"
	db "github.com/TF2Stadium/Helen/database"
)

type SlotDetails struct {
	Filled bool          `json:"filled"`
	Player PlayerSummary `json:"player,omitempty"`
	Ready  bool          `json:"ready"`
	InGame bool          `json:"ingame"`
}

type ClassDetails struct {
	Blu   SlotDetails `json:"blu"`
	Class string      `json:"class"`
	Red   SlotDetails `json:"red"`
}

type SpecDetails struct {
	Name    string `json:"name,omitempty"`
	SteamID string `json:"steamid,omitempty"`
}

type LobbyData struct {
	ID         uint   `json:"id"`
	Type       string `json:"type"`
	Players    int    `json:"players"`
	Map        string `json:"map"`
	League     string `json:"league"`
	Region     string `json:"region"`
	Mumble     bool   `json:"mumbleRequired"`
	MaxPlayers int    `json:"maxPlayers"`

	Classes []ClassDetails `json:"classes"`

	Leader      PlayerSummary `json:"leader"`
	CreatedAt   int64         `json:"createdAt"`
	State       int           `json:"state"`
	WhitelistID int           `json:"whitelistId"`

	Spectators []SpecDetails `json:"spectators,omitempty"`
}

type LobbyListData struct {
	Lobbies []LobbyData `json:"lobbies,omitempty"`
}

type LobbyConnectData struct {
	ID   uint   `json:"id"`
	Time int64  `json:"time"`
	Pass string `json:"password"`

	Game struct {
		Host string `json:"host"`
	} `json:"game"`

	Mumble struct {
		Address  string `json:"address"`
		Port     string `json:"port"`
		Password string `json:"password"`
		Channel  string `json:"channel"`
	} `json:"mumble"`
}

type LobbyEvent struct {
	ID uint `json:"id"`
}

func decorateSlotDetails(lobby *Lobby, slot int, includeDetails bool) SlotDetails {
	playerId, err := lobby.GetPlayerIdBySlot(slot)
	j := SlotDetails{Filled: err == nil}

	if err == nil && includeDetails {
		var player Player
		db.DB.First(&player, playerId)
		db.DB.Preload("Stats").First(&player, player.ID)

		j.Player = DecoratePlayerSummary(&player)

		ready, _ := lobby.IsPlayerReady(&player)
		j.Ready = ready

		ingame, _ := lobby.IsPlayerInGame(&player)
		j.InGame = ingame
	}

	return j
}

func DecorateLobbyData(lobby *Lobby, includeDetails bool) LobbyData {
	lobbyJs := LobbyData{
		ID:      lobby.ID,
		Type:    FormatMap[lobby.Type],
		Players: lobby.GetPlayerNumber(),
		Map:     lobby.MapName,
		League:  lobby.League,
		Mumble:  lobby.Mumble,
	}

	var classList = TypeClassList[lobby.Type]

	classes := make([]ClassDetails, len(classList))
	lobbyJs.MaxPlayers = NumberOfClassesMap[lobby.Type] * 2

	for slot, className := range classList {
		class := ClassDetails{
			Red:   decorateSlotDetails(lobby, slot, includeDetails),
			Blu:   decorateSlotDetails(lobby, slot+NumberOfClassesMap[lobby.Type], includeDetails),
			Class: className,
		}

		classes[slot] = class
	}

	lobbyJs.Classes = classes

	if !includeDetails {
		return lobbyJs
	}

	var leader Player
	db.DB.Where("steam_id = ?", lobby.CreatedBySteamID).First(&leader)

	lobbyJs.Leader = DecoratePlayerSummary(&leader)
	lobbyJs.CreatedAt = lobby.CreatedAt.Unix()
	lobbyJs.State = int(lobby.State)
	lobbyJs.WhitelistID = lobby.Whitelist

	var specIDs []uint
	db.DB.Table("spectators_players_lobbies").Where("lobby_id = ?", lobby.ID).Pluck("player_id", &specIDs)

	spectators := make([]SpecDetails, len(specIDs))

	for i, spectatorID := range specIDs {
		specPlayer := &Player{}
		db.DB.First(specPlayer, spectatorID)

		specJs := SpecDetails{
			Name:    specPlayer.Name,
			SteamID: specPlayer.SteamId,
		}

		spectators[i] = specJs
	}

	lobbyJs.Spectators = spectators

	return lobbyJs
}

func DecorateLobbyListData(lobbies []Lobby) LobbyListData {
	if len(lobbies) == 0 {
		return LobbyListData{}
	}

	var lobbyList = make([]LobbyData, len(lobbies))

	for i, lobby := range lobbies {
		lobbyData := DecorateLobbyData(&lobby, false)
		lobbyList[i] = lobbyData
	}

	listObj := LobbyListData{lobbyList}

	return listObj
}

func DecorateLobbyConnect(lobby *Lobby) LobbyConnectData {
	l := LobbyConnectData{}
	l.ID = lobby.ID
	l.Time = lobby.CreatedAt.Unix()
	l.Pass = lobby.ServerInfo.ServerPassword

	l.Game.Host = lobby.ServerInfo.Host

	l.Mumble.Address = config.Constants.MumbleAddr
	l.Mumble.Port = config.Constants.MumblePort
	l.Mumble.Password = config.Constants.MumblePassword
	l.Mumble.Channel = "match" + strconv.FormatUint(uint64(lobby.ID), 10)

	return l
}

func DecorateLobbyJoin(lobby *Lobby) LobbyEvent {
	return LobbyEvent{lobby.ID}
}

func DecorateLobbyLeave(lobby *Lobby) LobbyEvent {
	return LobbyEvent{lobby.ID}
}

func DecorateLobbyClosed(lobby *Lobby) LobbyEvent {
	return LobbyEvent{lobby.ID}
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package lobby

import (
	"fmt"
	"strings"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models/lobby/format"
	"github.com/TF2Stadium/Helen/models/player"
)

type SlotDetails struct {
	Slot         int            `json:"slot"`
	Filled       bool           `json:"filled"`
	Player       *player.Player `json:"player,omitempty"`
	Ready        *bool          `json:"ready,omitempty"`
	InGame       *bool          `json:"ingame,omitempty"`
	Requirements *Requirement   `json:"requirements,omitempty"`
	Password     bool           `json:"password"`
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
	ID                uint   `json:"id"`
	Mode              string `json:"gamemode"`
	Type              string `json:"type"`
	Players           int    `json:"players"`
	Map               string `json:"map"`
	League            string `json:"league"`
	Mumble            bool   `json:"mumbleRequired"`
	MaxPlayers        int    `json:"maxPlayers"`
	TwitchChannel     string `json:"twitchChannel"`
	TwitchRestriction string `json:"twitchRestriction"`

	RegionLock bool   `json:"regionLock"`
	SteamGroup string `json:"steamGroup"`

	Region struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"region"`

	Classes []ClassDetails `json:"classes"`

	Leader      player.Player `json:"leader"`
	CreatedAt   int64         `json:"createdAt"`
	State       int           `json:"state"`
	WhitelistID string        `json:"whitelistId"`

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

type SubstituteData struct {
	LobbyID uint   `json:"id"`
	Format  string `json:"type"`
	MapName string `json:"map"`

	Region struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"region"`

	RegionLock bool   `json:"regionLock"`
	Mumble     bool   `json:"mumbleRequired"`
	Team       string `sql:"-" json:"team"`
	Class      string `sql:"-" json:"class"`

	TwitchChannel     string `json:"twitchChannel"`
	TwitchRestriction string `json:"twitchRestriction"`
	SteamGroup        string `json:"steamGroup"`
	Password          bool   `json:"password"`
}

type LobbyEvent struct {
	ID uint `json:"id"`
}

func decorateSlotDetails(lobby *Lobby, slot int, playerInfo bool) SlotDetails {
	playerId, err := lobby.GetPlayerIDBySlot(slot)
	needsSub := lobby.SlotNeedsSubstitute(slot)

	slotDetails := SlotDetails{Slot: slot, Filled: err == nil && !needsSub}

	if err == nil && playerInfo && !needsSub {
		p, _ := player.GetPlayerByID(playerId)
		p.SetPlayerSummary()

		slotDetails.Player = p

		ready, _ := lobby.IsPlayerReady(p)
		slotDetails.Ready = &ready

		ingame := lobby.IsPlayerInGame(p)
		slotDetails.InGame = &ingame
	}

	if lobby.HasSlotRequirement(slot) {
		req, _ := lobby.GetSlotRequirement(slot)
		if req != nil {
			slotDetails.Requirements = req
			slotDetails.Password = req.Password != ""
		}
	}

	return slotDetails
}

var (
	stateString = map[State]string{
		Waiting:    "Waiting For Players",
		InProgress: "Lobby in Progress",
		Ended:      "Lobby Ended",
	}

	formatMap = map[format.Format]string{
		format.Sixes:      "6s",
		format.Highlander: "Highlander",
		format.Fours:      "4v4",
		format.Ultiduo:    "Ultiduo",
		format.Bball:      "Bball",
		format.Debug:      "Debug",
	}
)

func DecorateLobbyData(lobby *Lobby, playerInfo bool) LobbyData {
	lobbyData := LobbyData{
		ID:                lobby.ID,
		Mode:              lobby.Mode,
		Type:              formatMap[lobby.Type],
		Players:           lobby.GetPlayerNumber(),
		Map:               lobby.MapName,
		League:            lobby.League,
		Mumble:            lobby.Mumble,
		TwitchChannel:     lobby.TwitchChannel,
		TwitchRestriction: lobby.TwitchRestriction.String(),
		RegionLock:        lobby.RegionLock,

		SteamGroup: lobby.PlayerWhitelist,
	}

	lobbyData.Region.Name = lobby.RegionName
	lobbyData.Region.Code = lobby.RegionCode

	classList := format.GetClasses(lobby.Type)

	classes := make([]ClassDetails, len(classList))
	lobbyData.MaxPlayers = format.NumberOfClassesMap[lobby.Type] * 2

	for slot, className := range classList {
		class := ClassDetails{
			Red:   decorateSlotDetails(lobby, slot, playerInfo),
			Blu:   decorateSlotDetails(lobby, slot+format.NumberOfClassesMap[lobby.Type], playerInfo),
			Class: className,
		}

		classes[slot] = class
	}

	lobbyData.Classes = classes
	lobbyData.WhitelistID = lobby.Whitelist

	if !playerInfo {
		return lobbyData
	}

	if lobby.CreatedBySteamID != "" { // == "" during tests
		leader, _ := player.GetPlayerBySteamID(lobby.CreatedBySteamID)
		leader.SetPlayerSummary()
		lobbyData.Leader = *leader
	}

	lobbyData.CreatedAt = lobby.CreatedAt.Unix()
	lobbyData.State = int(lobby.State)

	var specIDs []uint
	db.DB.Table("spectators_players_lobbies").Where("lobby_id = ?", lobby.ID).Pluck("player_id", &specIDs)

	spectators := make([]SpecDetails, len(specIDs))

	for i, spectatorID := range specIDs {
		specPlayer, _ := player.GetPlayerByID(spectatorID)

		specJs := SpecDetails{
			Name:    specPlayer.Alias(),
			SteamID: specPlayer.SteamID,
		}

		spectators[i] = specJs
	}

	lobbyData.Spectators = spectators

	return lobbyData
}

func (l LobbyData) Send() {
	broadcaster.SendMessageToRoom(fmt.Sprintf("%d_public", l.ID), "lobbyData", l)
}

func (l LobbyData) SendToPlayer(steamid string) {
	broadcaster.SendMessage(steamid, "lobbyData", l)
}

func DecorateLobbyListData(lobbies []*Lobby, playerInfo bool) []LobbyData {
	var lobbyList = make([]LobbyData, len(lobbies))

	for i, lobby := range lobbies {
		lobbyData := DecorateLobbyData(lobby, playerInfo)
		lobbyList[i] = lobbyData
	}

	return lobbyList
}

func DecorateLobbyConnect(lob *Lobby, player *player.Player, slot int) LobbyConnectData {
	l := LobbyConnectData{}
	l.ID = lob.ID
	l.Time = lob.CreatedAt.Unix()
	l.Pass = lob.ServerInfo.ServerPassword
	l.Game.Host = lob.ServerInfo.Host

	l.Mumble.Address = config.Constants.MumbleAddr
	l.Mumble.Password = player.MumbleAuthkey
	team, _, _ := format.GetSlotTeamClass(lob.Type, slot)
	l.Mumble.Channel = fmt.Sprintf("Lobby #%d/%s", lob.ID, strings.ToUpper(team))
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

func DecorateSubstitute(slot *LobbySlot) SubstituteData {
	lobby, _ := GetLobbyByID(slot.LobbyID)

	substitute := SubstituteData{
		LobbyID:       lobby.ID,
		Format:        formatMap[lobby.Type],
		MapName:       lobby.MapName,
		Mumble:        lobby.Mumble,
		TwitchChannel: lobby.TwitchChannel,
		SteamGroup:    lobby.PlayerWhitelist,
		RegionLock:    lobby.RegionLock,
	}

	req, _ := lobby.GetSlotRequirement(slot.Slot)
	if req != nil {
		substitute.Password = req.Password != ""
	}

	substitute.Region.Name = lobby.RegionName
	substitute.Region.Code = lobby.RegionCode
	substitute.Team, substitute.Class, _ = format.GetSlotTeamClass(lobby.Type, slot.Slot)

	return substitute
}

func DecorateSubstituteList() []SubstituteData {
	slots := []*LobbySlot{}
	subList := []SubstituteData{}

	db.DB.Model(&LobbySlot{}).Joins("INNER JOIN lobbies ON lobbies.id = lobby_slots.lobby_id").Where("lobby_slots.needs_sub = ? AND lobbies.state = ?", true, InProgress).Find(&slots)

	for _, slot := range slots {
		subList = append(subList, DecorateSubstitute(slot))
	}

	return subList
}

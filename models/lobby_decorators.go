// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"strconv"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/bitly/go-simplejson"
)

func decorateSlotDetails(lobby *Lobby, slot int, includeDetails bool) *simplejson.Json {
	j := simplejson.New()

	playerId, err := lobby.GetPlayerIdBySlot(slot)
	j.Set("filled", err == nil)
	if err == nil && includeDetails {
		var player Player
		db.DB.First(&player, playerId)

		j.Set("player", DecoratePlayerSummaryJson(&player))
		ready, _ := lobby.IsPlayerReady(&player)
		j.Set("ready", ready)
		ingame, _ := lobby.IsPlayerInGame(&player)
		j.Set("inGame", ingame)
	}

	return j
}

func DecorateLobbyDataJSON(lobby *Lobby, includeDetails bool) *simplejson.Json {
	lobbyJs := simplejson.New()
	lobbyJs.Set("id", lobby.ID)
	lobbyJs.Set("type", FormatMap[lobby.Type])
	lobbyJs.Set("players", lobby.GetPlayerNumber())
	lobbyJs.Set("map", lobby.MapName)
	lobbyJs.Set("league", lobby.League)
	lobbyJs.Set("mumbleRequired", lobby.Mumble)

	var classes []*simplejson.Json

	var classList = TypeClassList[lobby.Type]
	lobbyJs.Set("maxPlayers", len(classList)*2)

	for slot, className := range classList {
		class := simplejson.New()

		class.Set("red", decorateSlotDetails(lobby, slot, includeDetails))
		class.Set("blu", decorateSlotDetails(lobby, slot+int(lobby.Type), includeDetails))
		class.Set("class", className)
		classes = append(classes, class)
	}
	lobbyJs.Set("classes", classes)

	if !includeDetails {
		return lobbyJs
	}

	var leader Player
	db.DB.Where("steam_id = ?", lobby.CreatedBySteamID).First(&leader)
	lobbyJs.Set("leader", DecoratePlayerSummaryJson(&leader))
	lobbyJs.Set("createdAt", lobby.CreatedAt.Unix())
	lobbyJs.Set("state", lobby.State)
	lobbyJs.Set("whitelistId", lobby.Whitelist)

	var spectators []*simplejson.Json
	for _, spectator := range lobby.Spectators {
		specJs := simplejson.New()
		specJs.Set("name", spectator.Name)
		specJs.Set("steamid", spectator.SteamId)
		spectators = append(spectators, specJs)
	}
	lobbyJs.Set("spectators", spectators)

	return lobbyJs
}

func DecorateLobbyListData(lobbies []Lobby) (string, error) {

	if len(lobbies) == 0 {
		return "{}", nil
	}

	var lobbyList []*simplejson.Json

	for _, lobby := range lobbies {
		lobbyJs := DecorateLobbyDataJSON(&lobby, false)
		lobbyList = append(lobbyList, lobbyJs)
	}

	listObj := simplejson.New()
	listObj.Set("lobbies", lobbyList)

	bytes, _ := listObj.MarshalJSON()
	return string(bytes), nil
}

func DecorateLobbyConnectJSON(lobby *Lobby) *simplejson.Json {
	json := simplejson.New()

	json.Set("id", lobby.ID)
	json.Set("time", lobby.CreatedAt.Unix())
	json.Set("password", lobby.ServerInfo.ServerPassword)

	game := simplejson.New()
	game.Set("host", lobby.ServerInfo.Host)
	json.Set("game", game)

	mumble := simplejson.New()
	mumble.Set("ip", "we still need to decide on mumble connections")
	mumble.Set("port", "we still need to decide on mumble connections")
	mumble.Set("channel", "match"+strconv.FormatUint(uint64(lobby.ID), 10))
	json.Set("mumble", mumble)

	return json
}

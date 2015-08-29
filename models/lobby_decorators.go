package models

import (
	"strconv"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/bitly/go-simplejson"
)

func decorateSlotDetails(lobby *Lobby, slot int) (string, string, bool) {
	steamid := ""
	name := ""
	ready := false

	playerId, err := lobby.GetPlayerIdBySlot(slot)
	if err == nil {
		var player Player
		db.DB.First(&player, playerId)

		steamid = player.SteamId
		name = player.Name
		ready, _ = lobby.IsPlayerReady(&player)
	}
	return steamid, name, ready
}

func DecorateLobbyDataJSON(lobby *Lobby) *simplejson.Json {
	lobbyJs := simplejson.New()
	lobbyJs.Set("id", lobby.ID)
	lobbyJs.Set("type", FormatMap[lobby.Type])
	lobbyJs.Set("createdAt", lobby.CreatedAt.Unix())
	lobbyJs.Set("players", lobby.GetPlayerNumber())
	lobbyJs.Set("map", lobby.MapName)
	var classes []*simplejson.Json

	var classList = LobbyFormatClassList(lobby.Type)
	lobbyJs.Set("maxPlayers", len(classList)*2)

	for slot, className := range classList {
		class := simplejson.New()
		red := simplejson.New()
		blu := simplejson.New()

		steamid, name, ready := decorateSlotDetails(lobby, slot)
		red.Set("steamid", steamid)
		red.Set("name", name)
		red.Set("ready", ready)

		steamid, name, ready = decorateSlotDetails(lobby, slot+TypePlayerCount[lobby.Type])
		blu.Set("steamid", steamid)
		blu.Set("name", name)
		blu.Set("ready", ready)

		class.Set("red", red)
		class.Set("blu", blu)
		class.Set("class", className)
		classes = append(classes, class)
	}
	lobbyJs.Set("classes", classes)

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
		lobbyJs := DecorateLobbyDataJSON(&lobby)
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

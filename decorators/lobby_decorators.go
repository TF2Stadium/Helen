package decorators

import (
	"strconv"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
)

func getSlotDetails(lobby *models.Lobby, slot int) (string, string, bool) {
	steamid := ""
	name := ""
	ready := false

	playerId, err := lobby.GetPlayerIdBySlot(slot)
	if err == nil {
		var player models.Player
		db.DB.First(&player, playerId)

		steamid = player.SteamId
		name = player.Name
		ready, _ = lobby.IsPlayerReady(&player)
	}
	return steamid, name, ready
}

func GetLobbyDataJSON(lobby models.Lobby) *simplejson.Json {
	lobbyJs := simplejson.New()
	lobbyJs.Set("id", lobby.ID)
	lobbyJs.Set("type", models.FormatMap[lobby.Type])
	lobbyJs.Set("createdAt", lobby.CreatedAt.Unix())
	lobbyJs.Set("players", lobby.GetPlayerNumber())
	lobbyJs.Set("map", lobby.MapName)
	classes := simplejson.New()

	var classMap = chelpers.FormatClassMap(lobby.Type)
	lobbyJs.Set("maxPlayers", len(classMap)*2)

	for className, slot := range classMap {
		class := simplejson.New()
		red := simplejson.New()
		blu := simplejson.New()

		steamid, name, ready := getSlotDetails(&lobby, slot)
		red.Set("steamid", steamid)
		red.Set("name", name)
		red.Set("ready", ready)

		steamid, name, ready = getSlotDetails(&lobby, slot+models.TypePlayerCount[lobby.Type])
		blu.Set("steamid", steamid)
		blu.Set("name", name)
		blu.Set("ready", ready)

		class.Set("red", red)
		class.Set("blu", blu)
		classes.Set(className, class)
	}
	lobbyJs.Set("classes", classes)

	return lobbyJs
}

func GetLobbyListData(lobbies []models.Lobby) (string, error) {

	if len(lobbies) == 0 {
		return "{}", nil
	}

	var lobbyList []*simplejson.Json

	for _, lobby := range lobbies {
		lobbyJs := GetLobbyDataJSON(lobby)
		lobbyList = append(lobbyList, lobbyJs)
	}

	listObj := simplejson.New()
	listObj.Set("lobbies", lobbyList)

	bytes, _ := listObj.MarshalJSON()
	return string(bytes), nil
}

func GetLobbyConnectJSON(lobby *models.Lobby) *simplejson.Json {
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

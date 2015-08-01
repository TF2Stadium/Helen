package decorators

import (
	chelpers "github.com/TF2Stadium/Server/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
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
	classes := simplejson.New()

	for className, slot := range chelpers.FormatClassMap(lobby.Type) {
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

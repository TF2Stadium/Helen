package decorators

import (
	"encoding/json"

	chelpers "github.com/TF2Stadium/Server/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
	"github.com/bitly/go-simplejson"
)

func getSlotDetails(lobby *models.Lobby, slot int) (string, string, bool) {
	steamid := ""
	name := ""
	ready := false

	if lobby.IsSlotFilled(slot) {
		var player *models.Player
		slot, _ := lobby.GetPlayerIdBySlot(slot)
		db.DB.First(&player, slot)

		steamid = player.SteamId
		name = player.Name
		ready, _ = lobby.IsPlayerReady(player)
	}
	return steamid, name, ready
}

func GetLobbyListData() (string, error) {
	count := 0
	db.DB.Where("state = ?", models.LobbyStateWaiting).Count(&count)

	if count == 0 {
		return "{}", nil
	}

	lobbyList := make([]*simplejson.Json, count)
	lobbies := make([]*models.Lobby, count)
	err := db.DB.Where("state = ?", models.LobbyStateWaiting).Find(&lobbies).Error

	if err != nil {
		return "{}", err
	}

	for lobbyIndex, lobby := range lobbies {
		lobbyJs := simplejson.New()
		lobbyJs.Set("id", lobby.ID)
		lobbyJs.Set("type", models.FormatMap[lobby.Type])
		lobbyJs.Set("createdAt", lobby.CreatedAt.String())
		lobbyJs.Set("players", lobby.GetPlayerNumber())
		classes := make([]*simplejson.Json, models.TypePlayerCount[lobby.Type])
		class := simplejson.New()

		for className, slot := range chelpers.FormatClassMap(lobby.Type) {
			players := simplejson.New()
			red := simplejson.New()
			blu := simplejson.New()

			steamid, name, ready := getSlotDetails(lobby, slot)
			red.Set("steamid", steamid)
			red.Set("name", name)
			red.Set("ready", ready)

			steamid, name, ready = getSlotDetails(lobby, slot+models.TypePlayerCount[lobby.Type])
			blu.Set("steamid", steamid)
			blu.Set("name", name)
			blu.Set("ready", ready)

			players.Set("red", red)
			players.Set("blu", blu)
			class.Set(className, players)
			classes[slot] = class
		}
		lobbyJs.Set("classes", classes)
		lobbyList[lobbyIndex] = lobbyJs
	}

	bytes, _ := json.Marshal(lobbyList)
	return string(bytes), nil
}

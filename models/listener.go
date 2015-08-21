package models

import (
	"strings"
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/bitly/go-simplejson"
)

var ticker *time.Ticker

func StartListener() {
	ticker = time.NewTicker(time.Second * 3)
	go listener()
	helpers.Logger.Debug("Listening for events on Pauling")
}

func listener() {
	for {
		select {
		case <-ticker.C:
			var jsonStr string
			Pauling.Call("Pauling.GetEvent", &Args{}, &jsonStr)
			event, _ := simplejson.NewFromReader(strings.NewReader(jsonStr))
			handleEvent(event)
		}
	}
}

func handleEvent(e *simplejson.Json) {
	event, err := e.Get("event").String()
	if err != nil { //event queue is empty
		return
	}

	switch event {
	case "playerDisc":
		slot := &LobbySlot{}
		lobbyid, _ := e.Get("lobbyId").Uint64()
		steamid, _ := e.Get("commId").String()
		player, _ := GetPlayerBySteamId(steamid)

		db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, uint(lobbyid)).First(slot)
		slot.InGame = false
		db.DB.Save(slot)
		/*TODO:
		 *Notify lobby channel with SendMessage(), but we can't import chelpers due to a
		 *import cycle
		 */

	case "playerConn":
		slot := &LobbySlot{}
		lobbyid, _ := e.Get("lobbyId").Uint64()
		steamid, _ := e.Get("commId").String()
		player, _ := GetPlayerBySteamId(steamid)

		err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, uint(lobbyid)).First(slot).Error
		if err == nil { //else, player isn't in the lobby, will be kicked by Pauling
			slot.InGame = true
			db.DB.Save(slot)
		}

	case "playerRep":
		lobbyid, _ := e.Get("lobbyId").Uint64()
		steamid, _ := e.Get("commId").String()
		player, _ := GetPlayerBySteamId(steamid)

		db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, uint(lobbyid)).Delete(&LobbySlot{})

	case "discFromServer":
		lobbyid, _ := e.Get("lobbyId").Uint64()

		lobby, _ := GetLobbyById(uint(lobbyid))
		lobby.Close(false)

	case "matchEnded":
		lobbyid, _ := e.Get("lobbyId").Uint64()

		lobby, _ := GetLobbyById(uint(lobbyid))
		lobby.Close(false)
	}
}

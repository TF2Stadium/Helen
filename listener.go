package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TF2Stadium/Helen/controllers/socket"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
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
			models.Pauling.Call("Pauling.GetEvent", &models.Args{}, &jsonStr)
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
		slot := &models.LobbySlot{}
		lobbyid, _ := e.Get("lobbyId").Uint64()
		steamid, _ := e.Get("commId").String()
		player, _ := models.GetPlayerBySteamId(steamid)

		db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, uint(lobbyid)).First(slot)
		slot.InGame = false
		db.DB.Save(slot)
		socket.SendMessageToRoom(strconv.FormatUint(uint64(lobbyid), 10),
			"sendNotification", fmt.Sprintf("%s has disconected from the server .",
				player.Name))

	case "playerConn":
		slot := &models.LobbySlot{}
		lobbyid, _ := e.Get("lobbyId").Uint64()
		steamid, _ := e.Get("commId").String()
		player, _ := models.GetPlayerBySteamId(steamid)

		err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, uint(lobbyid)).First(slot).Error
		if err == nil { //else, player isn't in the lobby, will be kicked by Pauling
			slot.InGame = true
			db.DB.Save(slot)
		}

	case "playerRep":
		lobbyid, _ := e.Get("lobbyId").Uint64()
		steamid, _ := e.Get("commId").String()
		player, _ := models.GetPlayerBySteamId(steamid)

		db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, uint(lobbyid)).Delete(&models.LobbySlot{})
		socket.SendMessageToRoom(strconv.FormatUint(uint64(lobbyid), 10),
			"sendNotification", fmt.Sprintf("%s has been reported.",
				player.Name))

	case "discFromServer":
		lobbyid, _ := e.Get("lobbyId").Uint64()

		lobby, _ := models.GetLobbyById(uint(lobbyid))
		lobby.Close(false)
		socket.SendMessageToRoom(strconv.FormatUint(uint64(lobbyid), 10),
			"sendNotification", "Disconnected from Server.")

	case "matchEnded":
		lobbyid, _ := e.Get("lobbyId").Uint64()

		lobby, _ := models.GetLobbyById(uint(lobbyid))
		lobby.Close(false)
		socket.SendMessageToRoom(strconv.FormatUint(uint64(lobbyid), 10),
			"sendNotification", "Lobby Ended.")

	}
}

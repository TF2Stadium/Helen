// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package main

import (
	"fmt"
	"net/rpc"
	"strconv"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/PlayerStatsScraper/steamid"
)

var ticker *time.Ticker

func StartListener() {
	if config.Constants.ServerMockUp {
		return
	}
	ticker = time.NewTicker(time.Second * 3)
	go listener()
	helpers.Logger.Debug("Listening for events on Pauling")
}

func listener() {
	for {
		select {
		case <-ticker.C:
			event := make(models.Event)
			err := models.Pauling.Call("Pauling.GetEvent", &models.Args{}, &event)

			if err == rpc.ErrShutdown { //Pauling has crashed
				//TODO
			} else if err != nil {
				helpers.Logger.Fatal(err)
			}
			if _, ok := event["empty"]; ok {
				handleEvent(event)
			}
		}
	}
}

func handleEvent(event map[string]interface{}) {
	switch event["name"] {
	case "playerDisc":
		slot := &models.LobbySlot{}
		lobbyid := event["lobbyId"].(uint)
		commId := event["commId"].(string)

		steamId, _ := steamid.CommIdToSteamId(commId)
		player, _ := models.GetPlayerBySteamId(steamId)

		db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).First(slot)
		slot.InGame = false
		db.DB.Save(slot)
		broadcaster.SendMessageToRoom(strconv.FormatUint(uint64(lobbyid), 10),
			"sendNotification", fmt.Sprintf("%s has disconected from the server .",
				player.Name))

	case "playerConn":
		slot := &models.LobbySlot{}
		lobbyid := event["lobbyId"].(uint)
		commId := event["commId"].(string)

		steamId, _ := steamid.CommIdToSteamId(commId)
		player, _ := models.GetPlayerBySteamId(steamId)
		err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).First(slot).Error
		if err == nil { //else, player isn't in the lobby, will be kicked by Pauling
			slot.InGame = true
			db.DB.Save(slot)
		}

	case "playerRep":
		lobbyid := event["lobbyId"].(uint)
		commId := event["commId"].(string)

		steamId, _ := steamid.CommIdToSteamId(commId)
		player, _ := models.GetPlayerBySteamId(steamId)

		db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).Delete(&models.LobbySlot{})
		broadcaster.SendMessageToRoom(strconv.FormatUint(uint64(lobbyid), 10),
			"sendNotification", fmt.Sprintf("%s has been reported.",
				player.Name))

	case "discFromServer":
		lobbyid := event["lobbyId"].(uint)

		lobby, _ := models.GetLobbyById(lobbyid)
		lobby.Close(false)
		broadcaster.SendMessageToRoom(strconv.FormatUint(uint64(lobbyid), 10),
			"sendNotification", "Disconnected from Server.")

	case "matchEnded":
		lobbyid := event["lobbyId"].(uint)

		lobby, _ := models.GetLobbyById(lobbyid)
		lobby.Close(false)
		broadcaster.SendMessageToRoom(strconv.FormatUint(uint64(lobbyid), 10),
			"sendNotification", "Lobby Ended.")

	case "getServers":
		var lobbies []*models.Lobby
		var activeStates = []models.LobbyState{models.LobbyStateWaiting, models.LobbyStateInProgress}
		db.DB.Where("lobby_state IN (?)", activeStates).Find(&lobbies)
		for _, lobby := range lobbies {
			info := models.ServerBootstrap{
				LobbyId: lobby.ID,
				Info:    lobby.ServerInfo,
			}
			for _, player := range lobby.BannedPlayers {
				commId, _ := steamid.SteamIdToCommId(player.SteamId)
				info.BannedPlayers = append(info.BannedPlayers, commId)
			}
			for _, slot := range lobby.Slots {
				var player *models.Player
				db.DB.Find(player, slot.PlayerId)
				commId, _ := steamid.SteamIdToCommId(player.SteamId)
				info.Players = append(info.Players, commId)
			}
			models.Pauling.Call("Pauling.SetupVerifier", &info, &struct{}{})
		}
	}
}

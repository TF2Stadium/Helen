// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package main

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

var ticker *time.Ticker

func StartListener() {
	if config.Constants.ServerMockUp {
		return
	}
	ticker = time.NewTicker(time.Millisecond * 500)
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
			if _, empty := event["empty"]; !empty {
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
		steamId := event["steamId"].(string)

		player, _ := models.GetPlayerBySteamId(steamId)

		db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).First(slot)
		helpers.LockRecord(slot.ID, slot)
		slot.InGame = false
		db.DB.Save(slot)
		helpers.UnlockRecord(slot.ID, slot)
		room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
		broadcaster.SendMessageToRoom(room,
			"sendNotification", fmt.Sprintf("%s has disconected from the server .",
				player.Name))
		go func() {
			t := time.After(time.Minute * 2)
			<-t
			lobby, _ := models.GetLobbyById(lobbyid)
			slot := &models.LobbySlot{}
			db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).First(slot)
			if !slot.InGame {
				helpers.LockRecord(lobby.ID, lobby)
				defer helpers.UnlockRecord(lobby.ID, lobby)
				lobby.RemovePlayer(player)
				broadcaster.SendMessage(player.SteamId, "sendNotification",
					"You have been removed from the lobby.")
			}

		}()

	case "playerConn":
		slot := &models.LobbySlot{}
		lobbyid := event["lobbyId"].(uint)
		steamId := event["steamId"].(string)

		player, _ := models.GetPlayerBySteamId(steamId)
		err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).First(slot).Error
		if err == nil { //else, player isn't in the lobby, will be kicked by Pauling
			helpers.LockRecord(slot.ID, slot)
			slot.InGame = true
			db.DB.Save(slot)
			helpers.UnlockRecord(slot.ID, slot)
		}

	case "playerRep":
		lobbyid := event["lobbyId"].(uint)
		steamId := event["steamId"].(string)

		player, _ := models.GetPlayerBySteamId(steamId)

		db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).Delete(&models.LobbySlot{})
		room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
		broadcaster.SendMessageToRoom(room,
			"sendNotification", fmt.Sprintf("%s has been reported.",
				player.Name))

	case "discFromServer":
		lobbyid := event["lobbyId"].(uint)

		lobby, _ := models.GetLobbyById(lobbyid)
		helpers.LockRecord(lobby.ID, lobby)
		lobby.Close(false)
		helpers.UnlockRecord(lobby.ID, lobby)
		room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
		broadcaster.SendMessageToRoom(room,
			"sendNotification", "Disconnected from Server.")

	case "matchEnded":
		lobbyid := event["lobbyId"].(uint)

		lobby, _ := models.GetLobbyById(lobbyid)
		helpers.LockRecord(lobby.ID, lobby)
		lobby.UpdateStats()
		lobby.Close(false)
		helpers.UnlockRecord(lobby.ID, lobby)
		room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
		broadcaster.SendMessageToRoom(room,
			"sendNotification", "Lobby Ended.")

	case "getServers":
		var lobbies []*models.Lobby
		var activeStates = []models.LobbyState{models.LobbyStateWaiting, models.LobbyStateInProgress}
		db.DB.Model(&models.Lobby{}).Where("state IN (?)", activeStates).Find(&lobbies)
		for _, lobby := range lobbies {
			info := models.ServerBootstrap{
				LobbyId: lobby.ID,
				Info:    lobby.ServerInfo,
			}
			for _, player := range lobby.BannedPlayers {
				info.BannedPlayers = append(info.BannedPlayers, player.SteamId)
			}
			for _, slot := range lobby.Slots {
				var player *models.Player
				db.DB.Find(player, slot.PlayerId)
				info.Players = append(info.Players, player.SteamId)
			}
			models.Pauling.Call("Pauling.SetupVerifier", &info, &struct{}{})
		}
	}
}

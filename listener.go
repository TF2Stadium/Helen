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

var eventChanMap = make(map[string](chan map[string]interface{}))

func StartListener() {
	if config.Constants.ServerMockUp {
		return
	}

	var events = [...]string{"test", "playerDisc", "playerConn", "discFromServer",
		"matchEnded", "playerRep", "playerSub"}

	for _, e := range events {
		eventChanMap[e] = make(chan map[string]interface{})
	}

	var ticker *time.Ticker
	ticker = time.NewTicker(time.Millisecond * 500)

	go eventListener()
	go listener(ticker)
	helpers.Logger.Debug("Listening for events on Pauling")
}

func listener(ticker *time.Ticker) {
	for {

		<-ticker.C

		event := make(models.Event)
		err := models.Pauling.Call("Pauling.GetEvent", &models.Args{}, &event)

		if err != nil {
			if err == rpc.ErrShutdown {
				models.PaulingReconnect()
				continue
			}
			helpers.Logger.Fatal(err)
		}
		if _, empty := event["empty"]; !empty {
			eventChanMap[event["name"].(string)] <- event
		}
	}

}

func eventListener() {
	for {
		select {
		case event := <-eventChanMap["playerDisc"]:
			lobbyid := event["lobbyId"].(uint)
			steamId := event["steamId"].(string)

			player, _ := models.GetPlayerBySteamId(steamId)
			lobby, _ := models.GetLobbyById(lobbyid)

			helpers.LockRecord(lobby.ID, lobby)
			lobby.SetNotInGame(player)
			helpers.UnlockRecord(lobby.ID, lobby)

			helpers.Logger.Debug("#%d, player %s<%s> disconnected",
				lobby.ID, player.Name, player.SteamId)

			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
			broadcaster.SendMessageToRoom(room,
				"sendNotification",
				fmt.Sprintf(`{"notification": "%s has disconected from the server."}`,
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
					broadcaster.SendMessage(player.SteamId,
						"sendNotification",
						"You have been removed from the lobby (Absence for >2 minutes).")
					helpers.Logger.Debug("#%d: %s<%s> removed")
				}

			}()

		case event := <-eventChanMap["playerConn"]:
			lobbyid := event["lobbyId"].(uint)
			steamId := event["steamId"].(string)

			player, _ := models.GetPlayerBySteamId(steamId)
			lobby, _ := models.GetLobbyById(lobbyid)

			helpers.LockRecord(lobby.ID, lobby)
			lobby.SetInGame(player)
			helpers.UnlockRecord(lobby.ID, lobby)

		case event := <-eventChanMap["playerRep"]:
			lobbyid := event["lobbyId"].(uint)
			steamId := event["steamId"].(string)

			player, _ := models.GetPlayerBySteamId(steamId)

			var slot *models.LobbySlot
			db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).Find(slot)
			slot.NeedSub = true
			db.DB.Save(slot)
			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
			broadcaster.SendMessageToRoom(room,
				"sendNotification",
				fmt.Sprintf(`{"notification": "%s has been reported."}`,
					player.Name))

			helpers.Logger.Debug("#%d: Reported player %s<%s>",
				lobbyid, player.Name, player.SteamId)

		case event := <-eventChanMap["discFromServer"]:
			lobbyid := event["lobbyId"].(uint)

			lobby, _ := models.GetLobbyById(lobbyid)

			helpers.Logger.Debug("#%d: Lost connection to %s", lobby.ID, lobby.ServerInfo.Host)

			helpers.LockRecord(lobby.ID, lobby)
			lobby.Close(false)
			helpers.UnlockRecord(lobby.ID, lobby)
			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
			broadcaster.SendMessageToRoom(room,
				"sendNotification", `{"notification": "Lobby Closed (Connection to server lost)."}`)

		case event := <-eventChanMap["matchEnded"]:
			lobbyid := event["lobbyId"].(uint)

			lobby, _ := models.GetLobbyById(lobbyid)

			helpers.Logger.Debug("#%d: Match Ended", lobbyid)

			helpers.LockRecord(lobby.ID, lobby)
			lobby.UpdateStats()
			lobby.Close(false)
			helpers.UnlockRecord(lobby.ID, lobby)
			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
			broadcaster.SendMessageToRoom(room,
				"sendNotification", `{"notification": ""Lobby Ended."}`)

		case <-eventChanMap["getServers"]:
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
}

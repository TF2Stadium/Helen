// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package main

import (
	"fmt"
	"io"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

func StartPaulingListener() {
	if config.Constants.ServerMockUp {
		return
	}

	var eventChanMap = make(map[string](chan models.Event))
	var events = [...]string{"test", "playerDisc", "playerConn", "discFromServer",
		"matchEnded", "playerSub"}

	for _, e := range events {
		eventChanMap[e] = make(chan models.Event)
	}

	go eventListener(eventChanMap)
	go listener(eventChanMap)
	helpers.Logger.Debug("Listening for events on Pauling")
}

func listener(eventChanMap map[string](chan models.Event)) {
	for {
		event := make(models.Event)
		err := models.Pauling.Call("Pauling.GetEvent", &models.Args{}, &event)

		if err != nil {
			if err == io.ErrUnexpectedEOF {
				models.PaulingReconnect()
				continue
			}
			helpers.Logger.Fatal(err)
		}
		eventChanMap[event["name"].(string)] <- event
	}
}

func eventListener(eventChanMap map[string](chan models.Event)) {
	for {
		select {
		case event := <-eventChanMap["playerDisc"]:
			lobbyid := event["lobbyId"].(uint)
			steamId := event["steamId"].(string)

			player, _ := models.GetPlayerBySteamId(steamId)
			lobby, _ := models.GetLobbyById(lobbyid)

			lobby.SetNotInGame(player)

			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
			broadcaster.SendMessageToRoom(room,
				"sendNotification",
				fmt.Sprintf(`{"notification": "%s has disconected from the server."}`,
					player.Name))

			time.AfterFunc(time.Second*10, func() {
				ingame, err := lobby.IsPlayerInGame(player)
				if err != nil {
					helpers.Logger.Error(err.Error())
				}
				if !ingame {
					sub, _ := models.NewSub(lobby.ID, player.SteamId)
					db.DB.Save(sub)
					models.BroadcastSubList()
					lobby.RemovePlayer(player)
				}
			})

		case event := <-eventChanMap["playerConn"]:
			lobbyid := event["lobbyId"].(uint)
			steamId := event["steamId"].(string)

			player, _ := models.GetPlayerBySteamId(steamId)
			lobby, _ := models.GetLobbyById(lobbyid)

			lobby.SetInGame(player)

		case event := <-eventChanMap["playerSub"]:
			lobbyid := event["lobbyId"].(uint)
			steamId := event["steamId"].(string)

			sub, err := models.NewSub(lobbyid, steamId)
			if err != nil {
				helpers.Logger.Error(err.Error())
				continue
			}
			db.DB.Save(sub)

			models.BroadcastSubList()

			player, _ := models.GetPlayerBySteamId(steamId)
			lobby, _ := models.GetLobbyById(lobbyid)
			lobby.RemovePlayer(player)

			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
			broadcaster.SendMessageToRoom(room,
				"sendNotification",
				fmt.Sprintf(`{"notification": "%s has been reported."}`,
					player.Name))

			//helpers.Logger.Debug("#%d: Reported player %s<%s>",
			//	lobbyid, player.Name, player.SteamId)

		case event := <-eventChanMap["discFromServer"]:
			lobbyid := event["lobbyId"].(uint)

			lobby, _ := models.GetLobbyByIdServer(lobbyid)

			helpers.Logger.Debug("#%d: Lost connection to %s", lobby.ID, lobby.ServerInfo.Host)

			lobby.Close(false)
			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
			broadcaster.SendMessageToRoom(room,
				"sendNotification", `{"notification": "Lobby Closed (Connection to server lost)."}`)

		case event := <-eventChanMap["matchEnded"]:
			lobbyid := event["lobbyId"].(uint)

			lobby, _ := models.GetLobbyByIdServer(lobbyid)

			helpers.Logger.Debug("#%d: Match Ended", lobbyid)

			lobby.UpdateStats()
			lobby.Close(false)
			room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobbyid))
			broadcaster.SendMessageToRoom(room,
				"sendNotification", `{"notification": ""Lobby Ended."}`)

		case <-eventChanMap["getServers"]:
			var lobbies []*models.Lobby
			var activeStates = []models.LobbyState{models.LobbyStateWaiting, models.LobbyStateInProgress}
			db.DB.Preload("ServerInfo").Model(&models.Lobby{}).Where("state IN (?)", activeStates).Find(&lobbies)
			for _, lobby := range lobbies {
				info := models.ServerBootstrap{
					LobbyId: lobby.ID,
					Info:    lobby.ServerInfo,
				}
				for _, player := range lobby.BannedPlayers {
					info.BannedPlayers = append(info.BannedPlayers, player.SteamId)
				}
				for _, slot := range lobby.Slots {
					var player = &models.Player{}
					db.DB.Find(player, slot.PlayerId)
					info.Players = append(info.Players, player.SteamId)
				}
				models.Pauling.Call("Pauling.SetupVerifier", &info, &struct{}{})
			}
		}
	}
}

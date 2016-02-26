package event

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

//Mirrored across github.com/Pauling/server
type Event struct {
	Name    string
	SteamID string

	LobbyID uint
	LogsID  int //logs.tf ID
}

//Event names
const (
	PlayerDisconnected string = "playerDisc"
	PlayerSubstituted  string = "playerSub"
	PlayerConnected    string = "playerConn"
	PlayerChat         string = "playerChat"

	DisconnectedFromServer string = "discFromServer"
	MatchEnded             string = "matchEnded"
	Test                   string = "test"
)

func StartListening() {
	q, err := helpers.AMQPChannel.QueueDeclare(config.Constants.RabbitMQQueue, false, false, false, false, nil)
	if err != nil {
		logrus.Fatal("Cannot declare queue ", err)
	}

	msgs, err := helpers.AMQPChannel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal("Cannot consume messages ", err)
	}

	go func() {
		for msg := range msgs {
			var event Event

			err := json.Unmarshal(msg.Body, &event)
			if err != nil {
				logrus.Error(err)
				continue
			}
			switch event.Name {
			case PlayerDisconnected:
				playerDisc(event.SteamID, event.LobbyID)
			case PlayerSubstituted:
				playerSub(event.SteamID, event.LobbyID)
			case PlayerConnected:
				playerConn(event.SteamID, event.LobbyID)
			case DisconnectedFromServer:
				disconnectedFromServer(event.LobbyID)
			case MatchEnded:
				matchEnded(event.LobbyID, event.LogsID)

			}
		}
	}()
}

func playerDisc(steamID string, lobbyID uint) {
	player, _ := models.GetPlayerBySteamID(steamID)
	lobby, _ := models.GetLobbyByID(lobbyID)

	lobby.SetNotInGame(player)

	models.SendNotification(fmt.Sprintf("%s has disconected from the server.", player.Alias()), int(lobby.ID))
	time.AfterFunc(time.Minute*2, func() {
		ingame, err := lobby.IsPlayerInGame(player)
		if err != nil {
			logrus.Error(err.Error())
		}
		if !ingame && lobby.CurrentState() != models.LobbyStateEnded {
			lobby.Substitute(player)
		}
	})
}

func playerConn(steamID string, lobbyID uint) {
	player, _ := models.GetPlayerBySteamID(steamID)
	lobby, _ := models.GetLobbyByID(lobbyID)

	lobby.SetInGame(player)
	models.SendNotification(fmt.Sprintf("%s has connected to the server.", player.Alias()), int(lobby.ID))
}

func playerSub(steamID string, lobbyID uint) {
	player, _ := models.GetPlayerBySteamID(steamID)
	lobby, _ := models.GetLobbyByID(lobbyID)
	lobby.Substitute(player)

	models.SendNotification(fmt.Sprintf("%s has been reported.", player.Alias()), int(lobby.ID))
}

func playerChat(lobbyID uint, steamID string, message string) {
	lobby, _ := models.GetLobbyByIDServer(lobbyID)
	player, _ := models.GetPlayerBySteamID(steamID)

	chatMessage := models.NewInGameChatMessage(lobby, player, message)
	chatMessage.Save()
	chatMessage.Send()
}

func disconnectedFromServer(lobbyID uint) {
	lobby, err := models.GetLobbyByIDServer(lobbyID)
	if err != nil {
		logrus.Error("Couldn't find lobby ", lobbyID, " in database")
		return
	}

	lobby.Close(false)
	models.SendNotification("Lobby Closed (Connection to server lost)", int(lobby.ID))
}

func matchEnded(lobbyID uint, logsID int) {
	lobby, _ := models.GetLobbyByIDServer(lobbyID)

	lobby.UpdateStats()
	lobby.Close(false)

	msg := fmt.Sprintf("Lobby Ended. Logs: http://logs.tf/%d", logsID)
	models.SendNotification(msg, int(lobby.ID))
}

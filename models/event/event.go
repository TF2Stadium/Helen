package event

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/chat"
	lobbypackage "github.com/TF2Stadium/Helen/models/lobby"
	playerpackage "github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/PlayerStatsScraper/steamid"
	"github.com/TF2Stadium/TF2RconWrapper"
)

//Mirrored across github.com/Pauling/server
type Event struct {
	Name     string
	SteamID  string
	PlayerID uint32 // used by fumble

	LobbyID    uint
	LogsID     int //logs.tf ID
	ClassTimes map[string]*classTime
	Players    []TF2RconWrapper.Player

	Self bool // true if
}

type classTime struct {
	Scout    time.Duration
	Soldier  time.Duration
	Pyro     time.Duration
	Demoman  time.Duration
	Heavy    time.Duration
	Engineer time.Duration
	Sniper   time.Duration
	Medic    time.Duration
	Spy      time.Duration
}

//Event names
const (
	PlayerDisconnected string = "playerDisc"
	PlayerSubstituted  string = "playerSub"
	PlayerConnected    string = "playerConn"
	PlayerChat         string = "playerChat"
	PlayerMumbleJoined string = "playerMumbleJoined"
	PlayerMumbleLeft   string = "playerMumbleLeft"
	PlayersList        string = "playersList"

	DisconnectedFromServer string = "discFromServer"
	MatchEnded             string = "matchEnded"
	Test                   string = "test"

	ReservationOver string = "reservationOver"
)

var stop = make(chan struct{})

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
		for {
			select {
			case msg := <-msgs:
				var event Event

				err := json.Unmarshal(msg.Body, &event)
				if err != nil {
					logrus.Fatal(err)
				}
				switch event.Name {
				case PlayerDisconnected:
					playerDisc(event.SteamID, event.LobbyID)
				case PlayerSubstituted:
					playerSub(event.SteamID, event.LobbyID, event.Self)
				case PlayerConnected:
					playerConn(event.SteamID, event.LobbyID)
				case DisconnectedFromServer:
					disconnectedFromServer(event.LobbyID)
				case MatchEnded:
					matchEnded(event.LobbyID, event.LogsID, event.ClassTimes)
				case ReservationOver:
					reservationEnded(event.LobbyID)
				case PlayerMumbleJoined:
					mumbleJoined(uint(event.PlayerID))
				case PlayerMumbleLeft:
					mumbleLeft(uint(event.PlayerID))
				case PlayersList:
					playersList(event.Players)
				}
			case <-stop:
				return
			}
		}
	}()
}

func StopListening() {
	stop <- struct{}{}
}

func reservationEnded(lobbyID uint) {
	lobby, _ := lobbypackage.GetLobbyByID(lobbyID)
	lobby.Close(false, false)
	chat.SendNotification("Lobby Closed (serveme.tf reservation ended)", int(lobby.ID))
}

func playerDisc(steamID string, lobbyID uint) {
	player, _ := playerpackage.GetPlayerBySteamID(steamID)
	lobby, _ := lobbypackage.GetLobbyByID(lobbyID)

	lobby.SetNotInGame(player)

	chat.SendNotification(fmt.Sprintf("%s has disconected from the server.", player.Alias()), int(lobby.ID))

	lobby.AfterPlayerNotInGameFunc(player, 5*time.Minute, func() {
		lobby.Substitute(player)
		player.NewReport(playerpackage.Substitute, lobby.ID)
		chat.SendNotification(fmt.Sprintf("%s has been reported for not joining the game in 5 minutes", player.Alias()), int(lobby.ID))
	})
}

func playerConn(steamID string, lobbyID uint) {
	player, _ := playerpackage.GetPlayerBySteamID(steamID)
	lobby, _ := lobbypackage.GetLobbyByID(lobbyID)

	lobby.SetInGame(player)
	chat.SendNotification(fmt.Sprintf("%s has connected to the server.", player.Alias()), int(lobby.ID))
}

func playerSub(steamID string, lobbyID uint, self bool) {
	player, _ := playerpackage.GetPlayerBySteamID(steamID)
	lobby, err := lobbypackage.GetLobbyByID(lobbyID)
	if err != nil {
		logrus.Error(err)
		return
	}

	lobby.Substitute(player)
	if self {
		player.NewReport(playerpackage.Substitute, lobby.ID)

	} else {
		// ban player from joining lobbies for 30 minutes
		player.NewReport(playerpackage.Vote, lobby.ID)
	}

	chat.SendNotification(fmt.Sprintf("%s has been reported.", player.Alias()), int(lobby.ID))
}

func playerChat(lobbyID uint, steamID string, message string) {
	lobby, _ := lobbypackage.GetLobbyByIDServer(lobbyID)
	player, _ := playerpackage.GetPlayerBySteamID(steamID)

	chatMessage := chat.NewInGameChatMessage(lobby.ID, player, message)
	chatMessage.Save()
	chatMessage.Send()
}

func disconnectedFromServer(lobbyID uint) {
	lobby, err := lobbypackage.GetLobbyByIDServer(lobbyID)
	if err != nil {
		logrus.Error("Couldn't find lobby ", lobbyID, " in database")
		return
	}

	lobby.Close(false, false)
	chat.SendNotification("Lobby Closed (Connection to server lost)", int(lobby.ID))
}

func matchEnded(lobbyID uint, logsID int, classTimes map[string]*classTime) {
	lobby, err := lobbypackage.GetLobbyByIDServer(lobbyID)
	if err != nil {
		logrus.Error(err)
		return
	}
	lobby.Close(false, true)

	logs := fmt.Sprintf("http://logs.tf/%d", logsID)
	msg := fmt.Sprintf("Lobby Ended. Logs: %s", logs)
	chat.SendNotification(msg, int(lobby.ID))

	room := fmt.Sprintf("%d_private", lobby.ID)
	broadcaster.SendMessageToRoom(room, "lobbyLogs", logs)
	lobby.UpdateHours(logsID)
}

func mumbleJoined(playerID uint) {
	player, _ := playerpackage.GetPlayerByID(playerID)
	id, _ := player.GetLobbyID(false)
	if id == 0 { // player joined mumble lobby for closed channel
		return
	}

	lobby, _ := lobbypackage.GetLobbyByID(id)
	lobby.SetInMumble(player)
}

func mumbleLeft(playerID uint) {
	player, _ := playerpackage.GetPlayerByID(playerID)
	id, _ := player.GetLobbyID(false)
	if id == 0 { // player joined mumble lobby for closed channel
		return
	}

	lobby, _ := lobbypackage.GetLobbyByID(id)
	lobby.SetNotInMumble(player)
}

func playersList(players []TF2RconWrapper.Player) {
	for _, player := range players {
		commid, _ := steamid.SteamIdToCommId(player.SteamID)
		player, err := playerpackage.GetPlayerBySteamID(commid)
		if err != nil {
			continue
		}

		id, _ := player.GetLobbyID(false)
		if id == 0 {
			continue
		}

		lobby, _ := lobbypackage.GetLobbyByID(id)
		if !lobby.IsPlayerInGame(player) {
			lobby.SetInGame(player)
		}
	}
}

package rpc

import (
	"fmt"
	"time"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

// Event represents an event triggered by Pauling
type Event struct {
	Name string

	LobbyID  uint
	PlayerID uint
	LogsID   int //logs.tf ID
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

// Handle event e
func (Event) Handle(e Event, nop *struct{}) error {
	switch e.Name {
	case PlayerDisconnected:
		playerDisc(e.PlayerID, e.LobbyID)
	case PlayerSubstituted:
		playerSub(e.PlayerID, e.LobbyID)
	case PlayerConnected:
		playerConn(e.PlayerID, e.LobbyID)
	case DisconnectedFromServer:
		disconnectedFromServer(e.LobbyID)
	case MatchEnded:
		matchEnded(e.LobbyID, e.LogsID)
	}

	return nil
}

func playerDisc(playerID, lobbyID uint) {
	player, _ := models.GetPlayerByID(playerID)
	lobby, _ := models.GetLobbyByID(lobbyID)

	lobby.SetNotInGame(player)

	models.SendNotification(fmt.Sprintf("%s has disconected from the server.", player.Name), int(lobby.ID))
	time.AfterFunc(time.Minute*2, func() {
		ingame, err := lobby.IsPlayerInGame(player)
		if err != nil {
			helpers.Logger.Error(err.Error())
		}
		if !ingame && lobby.CurrentState() != models.LobbyStateEnded {
			lobby.Substitute(player)
		}
	})
}

func playerConn(playerID, lobbyID uint) {
	player, _ := models.GetPlayerByID(playerID)
	lobby, _ := models.GetLobbyByID(lobbyID)

	lobby.SetInGame(player)
	models.SendNotification(fmt.Sprintf("%s has connected to the server.", player.Alias()), int(lobby.ID))
}

func playerSub(playerID, lobbyID uint) {
	player, _ := models.GetPlayerByID(playerID)
	lobby, _ := models.GetLobbyByID(lobbyID)
	lobby.Substitute(player)

	models.SendNotification(fmt.Sprintf("%s has been reported.", player.Name), int(lobby.ID))
}

func playerChat(lobbyID uint, playerID uint, message string) {
	lobby, _ := models.GetLobbyByIDServer(lobbyID)
	player, _ := models.GetPlayerByID(playerID)

	chatMessage := models.NewInGameChatMessage(lobby, player, message)
	chatMessage.Save()
	chatMessage.Send()
}

func disconnectedFromServer(lobbyID uint) {
	lobby, _ := models.GetLobbyByIDServer(lobbyID)

	helpers.Logger.Debug("#%d: Lost connection to %s", lobby.ID, lobby.ServerInfo.Host)

	lobby.Close(false)
	models.SendNotification("Lobby Closed (Connection to server lost)", int(lobby.ID))
}

func matchEnded(lobbyID uint, logsID int) {
	lobby, _ := models.GetLobbyByIDServer(lobbyID)

	helpers.Logger.Debug("#%d: Match Ended", lobbyID)

	lobby.UpdateStats()
	lobby.Close(false)

	msg := fmt.Sprintf("Lobby Ended. Logs: http://logs.tf/%d", logsID)
	models.SendNotification(msg, int(lobby.ID))
}

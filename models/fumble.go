package models

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/fumble/mumble"
)

func FumbleLobbyCreated(lob *Lobby) error {
	err := fumble.Call("Fumble.CreateLobby", lob.ID, &struct{}{})

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func fumbleAllowPlayer(lobbyId uint, playerName string, playerTeam string) error {
	user := mumble.User{}
	user.Name = playerName
	user.Team = mumble.Team(playerTeam)

	err := fumble.Call("Fumble.AllowPlayer", &mumble.LobbyArgs{
		User: user, LobbyID: lobbyId}, &struct{}{})
	if err != nil {
		logrus.Error(err)
	}

	return nil
}

func FumbleLobbyPlayerJoinedSub(lob *Lobby, player *Player, slot int) {
	team, class, _ := LobbyGetSlotInfoString(lob.Type, slot)
	fumbleAllowPlayer(lob.ID, strings.ToUpper(team)+"_"+strings.ToUpper(class), strings.ToUpper(team))
}

func FumbleLobbyPlayerJoined(lob *Lobby, player *Player, slot int) {
	team, class, _ := LobbyGetSlotInfoString(lob.Type, slot)
	fumbleAllowPlayer(lob.ID, strings.ToUpper(team)+"_"+strings.ToUpper(class), "")
}

func FumbleLobbyEnded(lob *Lobby) {
	err := fumble.Call("Fumble.EndLobby", lob.ID, nil)
	if err != nil {
		logrus.Error(err)
	}
}

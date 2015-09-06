package testhelpers

import "github.com/TF2Stadium/Helen/models"

func CreatePlayer() *models.Player {
	player, _ := models.NewPlayer(RandSeq(5))
	player.Save()
	return player
}

func CreateLobby() *models.Lobby {
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "etf2l", models.ServerRecord{}, 0)
	lobby.Save()
	return lobby
}

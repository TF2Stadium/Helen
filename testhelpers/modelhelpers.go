// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

func CreatePlayer() *models.Player {
	player, _ := models.NewPlayer(RandSeq(5))
	player.Save()
	return player
}

func CreatePlayerMod() *models.Player {
	p := CreatePlayer()
	p.Role = helpers.RoleMod
	p.Save()
	return p
}

func CreatePlayerAdmin() *models.Player {
	p := CreatePlayer()
	p.Role = helpers.RoleAdmin
	p.Save()
	return p
}

func CreateLobby() *models.Lobby {
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "etf2l", models.ServerRecord{}, 0)
	lobby.Save()
	return lobby
}

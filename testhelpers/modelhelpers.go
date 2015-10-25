// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"math/rand"
	"time"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

// taken from http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func CreatePlayer() *models.Player {
	player, _ := models.NewPlayer(RandSeq(4))
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
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "etf2l", models.ServerRecord{}, 0, false)
	lobby.Save()
	return lobby
}

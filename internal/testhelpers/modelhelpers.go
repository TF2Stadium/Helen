// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

func CreatePlayer() *models.Player {
	bytes := make([]byte, 10)
	rand.Read(bytes)
	steamID := base64.URLEncoding.EncodeToString(bytes)

	player, _ := models.NewPlayer(steamID)
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
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "etf2l", models.ServerRecord{}, "0", false, "", "")
	lobby.Save()
	return lobby
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/gameserver"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/lobby/format"
	"github.com/TF2Stadium/Helen/models/player"
)

func CreatePlayer() *player.Player {
	bytes := make([]byte, 10)
	rand.Read(bytes)
	steamID := base64.URLEncoding.EncodeToString(bytes)

	player, _ := player.NewPlayer(steamID)
	player.MumbleUsername = steamID
	player.Save()
	return player
}

func CreatePlayerMod() *player.Player {
	p := CreatePlayer()
	p.Role = helpers.RoleMod
	p.Save()
	return p
}

func CreatePlayerAdmin() *player.Player {
	p := CreatePlayer()
	p.Role = helpers.RoleAdmin
	p.Save()
	return p
}

func CreateLobby() *lobby.Lobby {
	lobby := lobby.NewLobby("cp_badlands", format.Sixes, "etf2l", gameserver.ServerRecord{}, "0", false, "")
	lobby.Save()
	lobby.CreateLock()
	return lobby
}

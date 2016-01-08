// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"strconv"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
)

// major ver -> migration routine
var migrationRoutines = map[uint64]func(){
	2: lobbyTypeChange,
}

func whitelist_id_string() {
	var count int

	db.DB.Table("lobbies").Count(&count)
	if count == 0 {
		db.DB.Exec("ALTER TABLE lobbies DROP COLUMN whitelist")
		db.DB.Exec("ALTER TABLE lobbies ADD whitelist varchar(255)")
	}

	var whitelistIDs []int
	var lobbyIDs []uint

	db.DB.Table("lobbies").Order("whitelist").Pluck("whitelist", &whitelistIDs)
	if len(whitelistIDs) == 0 {
		return
	}

	db.DB.Table("lobbies").Order("id").Pluck("id", &lobbyIDs)

	db.DB.Exec("ALTER TABLE lobbies DROP whitelist")
	db.DB.Exec("ALTER TABLE lobbies ADD whitelist varchar(255)")

	for i, lobbyID := range lobbyIDs {
		db.DB.Model(&models.Lobby{}).Where("id = ?", lobbyID).Update("whitelist", strconv.Itoa(whitelistIDs[i]))
	}
}

func lobbyTypeChange() {
	newLobbyType := map[int]models.LobbyType{
		6: models.LobbyTypeSixes,
		9: models.LobbyTypeHighlander,
		4: models.LobbyTypeFours,
		3: models.LobbyTypeUltiduo,
		2: models.LobbyTypeBball,
		1: models.LobbyTypeDebug,
	}

	var lobbyIDs []uint
	db.DB.Table("lobbies").Order("id").Pluck("id", &lobbyIDs)

	for _, lobbyID := range lobbyIDs {
		var old int
		db.DB.DB().QueryRow("SELECT type FROM lobbies WHERE id = $1", lobbyID).Scan(&old)
		db.DB.Table("lobbies").Where("id = ?", lobbyID).Update("type", newLobbyType[old])
	}
}

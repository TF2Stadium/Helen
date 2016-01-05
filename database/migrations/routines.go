// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"strconv"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
)

var migrationsRoutines = map[uint64]func(){}

func whitelist_id_string() {
	var count int

	db.DB.DB().Stats()
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

	db.DB.Model(&models.Lobby{}).DropColumn("whitelist")
	db.DB.Exec("ALTER TABLE lobbies ADD whitelist varchar(255)")

	for i, lobbyID := range lobbyIDs {
		db.DB.Model(&models.Lobby{}).Where("id = ?", lobbyID).Update("whitelist", strconv.Itoa(whitelistIDs[i]))
	}
}

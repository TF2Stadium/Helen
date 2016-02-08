// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"sync"

	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
)

var once = new(sync.Once)

func Do() {
	database.DB.Exec("CREATE EXTENSION IF NOT EXISTS hstore")
	database.DB.AutoMigrate(&models.Player{})
	database.DB.AutoMigrate(&models.Lobby{})
	database.DB.AutoMigrate(&models.LobbySlot{})
	database.DB.AutoMigrate(&models.ServerRecord{})
	database.DB.AutoMigrate(&models.PlayerStats{})
	database.DB.AutoMigrate(&models.AdminLogEntry{})
	database.DB.AutoMigrate(&models.PlayerBan{})

	database.DB.Model(&models.LobbySlot{}).AddUniqueIndex("idx_lobby_slot_lobby_id_slot", "lobby_id", "slot")
	database.DB.AutoMigrate(&models.ChatMessage{})
	database.DB.AutoMigrate(&models.Requirement{})
	database.DB.AutoMigrate(&Constant{})

	once.Do(checkSchema)
}

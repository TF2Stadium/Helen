// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
)

func Do() {
	database.DB.AutoMigrate(&models.Player{})
	database.DB.AutoMigrate(&models.Lobby{})
	database.DB.AutoMigrate(&models.LobbySlot{})
	database.DB.AutoMigrate(&models.ServerRecord{})
	database.DB.AutoMigrate(&models.PlayerStats{})
	database.DB.AutoMigrate(&models.PlayerSetting{})
	database.DB.AutoMigrate(&models.AdminLogEntry{})
	database.DB.AutoMigrate(&models.PlayerBan{})

	database.DB.Model(&models.LobbySlot{}).AddUniqueIndex("idx_lobby_slot_lobby_id_slot", "lobby_id", "slot")
	database.DB.Model(&models.PlayerSetting{}).AddUniqueIndex("idx_player_id_key", "player_id", "key")
}

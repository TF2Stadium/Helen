// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"sync"

	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/models/chat"
	"github.com/TF2Stadium/Helen/models/gameserver"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/player"
)

var once = new(sync.Once)

func Do() {
	database.DB.Exec("CREATE EXTENSION IF NOT EXISTS hstore")
	database.DB.AutoMigrate(&player.Player{})
	database.DB.AutoMigrate(&lobby.Lobby{})
	database.DB.AutoMigrate(&lobby.LobbySlot{})
	database.DB.AutoMigrate(&gameserver.ServerRecord{})
	database.DB.AutoMigrate(&player.PlayerStats{})
	database.DB.AutoMigrate(&models.AdminLogEntry{})
	database.DB.AutoMigrate(&player.PlayerBan{})
	database.DB.AutoMigrate(&chat.ChatMessage{})
	database.DB.AutoMigrate(&lobby.Requirement{})
	database.DB.AutoMigrate(&Constant{})
	database.DB.AutoMigrate(&gameserver.StoredServer{})
	database.DB.AutoMigrate(&player.Report{})

	database.DB.Model(&lobby.LobbySlot{}).
		AddUniqueIndex("idx_lobby_slot_lobby_id_slot", "lobby_id", "slot")
	database.DB.Model(&lobby.LobbySlot{}).
		AddUniqueIndex("idx_lobby_id_player_id", "lobby_id", "player_id")
	database.DB.Model(&lobby.LobbySlot{}).
		AddUniqueIndex("idx_requirement_lobby_id_slot", "lobby_id", "slot")
	database.DB.Model(&lobby.LobbySlot{}).
		AddForeignKey("player_id", "players(id)", "RESTRICT", "RESTRICT")

	checkSchema()
}

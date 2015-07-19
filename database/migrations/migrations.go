package migrations

import (
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
)

func Do() {
	database.DB.AutoMigrate(&models.Player{})
	database.DB.AutoMigrate(&models.Lobby{})
	database.DB.AutoMigrate(&models.LobbySlot{})

	database.DB.Model(&models.LobbySlot{}).AddUniqueIndex("idx_lobby_slot_lobby_id_slot", "lobby_id", "slot")
}

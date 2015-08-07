package migrations

import (
	"os"

	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
)

func Do() {
	database.DB.AutoMigrate(&models.Player{})
	database.DB.AutoMigrate(&models.Lobby{})
	database.DB.AutoMigrate(&models.LobbySlot{})
	database.DB.AutoMigrate(&models.ServerRecord{})
	database.DB.AutoMigrate(&models.PlayerStats{})
	database.DB.AutoMigrate(&models.PlayerSetting{})

	database.DB.Model(&models.LobbySlot{}).AddUniqueIndex("idx_lobby_slot_lobby_id_slot", "lobby_id", "slot")
	database.DB.Model(&models.PlayerSetting{}).AddUniqueIndex("idx_player_id_key", "player_id", "key")
}

func TestCleanup() {
	if os.Getenv("DEPLOYMENT_ENV") == "" {
		os.Setenv("DEPLOYMENT_ENV", "test")
		defer os.Unsetenv("DEPLOYMENT_ENV")
	}
	config.SetupConstants()
	database.Init()

	database.DB.Exec("DROP SCHEMA public CASCADE;")
	database.DB.Exec("CREATE SCHEMA public;")

	Do()
}

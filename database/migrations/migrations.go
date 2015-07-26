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

	database.DB.Model(&models.LobbySlot{}).AddUniqueIndex("idx_lobby_slot_lobby_id_slot", "lobby_id", "slot")
}

func TestCleanup() {
	if os.Getenv("DEPLOYMENT_ENV") == "" {
		os.Setenv("DEPLOYMENT_ENV", "test")
		defer os.Unsetenv("DEPLOYMENT_ENV")
	}
	config.SetupConstants()
	database.Init()

	database.DB.Exec("DROP TABLE lobbies;")
	database.DB.Exec("DROP TABLE players;")
	database.DB.Exec("DROP TABLE lobby_slots;")
	database.DB.Exec("DROP TABLE banned_players_lobbies;")
	database.DB.Exec("DROP TABLE spectators_players_lobbies;")

	Do()
}

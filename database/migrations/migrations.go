package migrations

import (
	"log"
	"os"

	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
)

func Do() {
	database.DB.AutoMigrate(&models.Player{})
	database.DB.AutoMigrate(&models.Lobby{})
	database.DB.AutoMigrate(&models.LobbySlot{})

	database.DB.Model(&models.LobbySlot{}).AddUniqueIndex("idx_lobby_slot_lobby_id_slot", "lobby_id", "slot")
}

func TestCleanup() {
	os.Setenv("DEPLOYMENT_ENV", "test")
	config.SetupConstants()
	os.Unsetenv("DEPLOYMENT_ENV")
	database.Init()

	database.DB.Exec("DROP TABLE lobbies;")
	database.DB.Exec("DROP TABLE players;")
	database.DB.Exec("DROP TABLE lobby_slots;")
	database.DB.Exec("DROP TABLE banned_players_lobbies;")
	database.DB.Exec("DROP TABLE spectators_players_lobbies;")

	log.Println(database.DB)

	Do()
}

package database

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/TF2Stadium/Server/config"
	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/database/migrations"
	"github.com/TF2Stadium/Server/models"
	"github.com/stretchr/testify/assert"
)

var steamid = "76561198074578368"

func cleanup() {
	config.SetupConstants()
	db.Test()
	fmt.Println("[Test.Database] IsTest? " + strconv.FormatBool(db.IsTest))
	db.Init()

	db.DB.Exec("DROP TABLE lobbies;")
	db.DB.Exec("DROP TABLE players;")

	migrations.Do()
}

func TestDatabasePing(t *testing.T) {
	cleanup()
	assert.Nil(t, db.DB.DB().Ping())
}

// test the creation of a player
func TestDatabaseSave(t *testing.T) {
	cleanup()
	player := models.NewPlayer(steamid)

	err := player.Save()
	assert.Equal(t, err, nil)
	assert.NotEqual(t, player.ID, 0)
	assert.Equal(t, player.SteamId, steamid)

	player.Name = "John"
	err = player.Save()
	assert.Equal(t, err, nil)
	assert.Equal(t, player.Name, "John")
}

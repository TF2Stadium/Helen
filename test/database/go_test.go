package database

import (
	"testing"

	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/database/migrations"
	"github.com/TF2Stadium/Server/models"
	"github.com/stretchr/testify/assert"
)

var steamid = "76561198074578368"

func TestDatabasePing(t *testing.T) {
	migrations.TestCleanup()
	assert.Nil(t, db.DB.DB().Ping())
}

// test the creation of a player
func TestDatabaseSave(t *testing.T) {
	migrations.TestCleanup()
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

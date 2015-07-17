package database

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

// should always be the 1st test (like a setup)
func TestDatabase(t *testing.T) {
	// start the database connection
	config.SetupConstants()
	database.Test()
	fmt.Println("[Test.Database] IsTest? " + strconv.FormatBool(database.IsTest))
	database.Init()

}

var steamid = "76561198074578368"

func cleanup() {
	database.Database.C("players").Remove(bson.M{"steamid": steamid})
}

// test the creation of a player
func TestDatabaseSave(t *testing.T) {
	cleanup()
	player := models.NewPlayer(steamid)

	err := player.Save()
	assert.Equal(t, err, nil)
	assert.NotEqual(t, player.Id, "")
	assert.Equal(t, player.SteamId, steamid)

	player.Name = "John"
	err = player.Save()
	assert.Equal(t, err, nil)
	assert.Equal(t, player.Name, "John")
}

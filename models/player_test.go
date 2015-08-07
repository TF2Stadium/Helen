package models_test

import (
	"testing"

	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/database/migrations"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/TF2Stadium/Server/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
}

func TestPlayerInfoFetching(t *testing.T) {
	migrations.TestCleanup()

	if config.Constants.SteamDevApiKey == "your steam dev api key" {
		return
	}

	// disable mock mode because we're actually testing it
	config.Constants.SteamApiMockUp = false

	player, playErr := models.NewPlayer("76561197999073985")
	assert.Nil(t, playErr)

	assert.Equal(t, "http://steamcommunity.com/id/nonagono/", player.Profileurl)
	assert.Regexp(t, "(.*)steamcommunity/public/images/avatars/(.*).jpg", player.Avatar)

	assert.True(t, player.GameHours >= 3000)

	player.Stats.PlayedCountSet(models.LobbyTypeSixes, 3)
	player.Stats.PlayedCountSet(models.LobbyTypeHighlander, 7)
	player.Stats.PlayedCountIncrease(models.LobbyTypeSixes) // sixes: 3 -> 4

	assert.Equal(t, 4, player.Stats.PlayedCountGet(models.LobbyTypeSixes))
	assert.Equal(t, 7, player.Stats.PlayedCountGet(models.LobbyTypeHighlander))

	database.DB.Save(player)

	player2, err := models.GetPlayerWithStats(player.SteamId)
	assert.Nil(t, err)

	assert.Equal(t, 4, player2.Stats.PlayedCountGet(models.LobbyTypeSixes))
	assert.Equal(t, 7, player2.Stats.PlayedCountGet(models.LobbyTypeHighlander))
	assert.Equal(t, "http://steamcommunity.com/id/nonagono/", player2.Profileurl)
}

func TestPlayerSettings(t *testing.T) {
	migrations.TestCleanup()

	player, _ := models.NewPlayer("76561197999073985")

	settings, err := player.GetSettings()

	assert.Nil(t, err)
	assert.Equal(t, 0, len(settings))

	err = player.SetSetting("foo", "bar")
	assert.Nil(t, err)

	settings, err = player.GetSettings()
	assert.Nil(t, err)
	assert.Equal(t, "foo", settings[0].Key)
	assert.Equal(t, "bar", settings[0].Value)

	setting, err := player.GetSetting("foo")
	assert.Nil(t, err)
	assert.Equal(t, "bar", setting.Value)

	err = player.SetSetting("hello", "world")
	assert.Nil(t, err)

	settings, err = player.GetSettings()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(settings))
}

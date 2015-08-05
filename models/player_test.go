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

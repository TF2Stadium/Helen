package models_test

import (
	"testing"

	"github.com/TF2Stadium/Server/models"
	"github.com/stretchr/testify/assert"
)

func TestLobbiesPlayed(t *testing.T) {
	player, playErr := models.NewPlayer("smurf")
	assert.Nil(t, playErr)

	player.Stats.LobbiesPlayed.Set(models.LobbyTypeSixes, 5)
	player.Stats.LobbiesPlayed.Set(models.LobbyTypeHighlander, 8)
	player.Stats.LobbiesPlayed.Increase(models.LobbyTypeSixes) // sixes: 5 -> 6

	assert.Equal(t, 6, player.Stats.LobbiesPlayed.Get(models.LobbyTypeSixes))
	assert.Equal(t, 8, player.Stats.LobbiesPlayed.Get(models.LobbyTypeHighlander))
	assert.Equal(t, "6,8", player.Stats.LobbiesPlayed.String())

	player.Stats.LobbiesPlayed.Data = "7,1"
	player.Stats.LobbiesPlayed.Parse()

	assert.Equal(t, 7, player.Stats.LobbiesPlayed.Get(models.LobbyTypeSixes))
	assert.Equal(t, 1, player.Stats.LobbiesPlayed.Get(models.LobbyTypeHighlander))
	assert.Equal(t, "7,1", player.Stats.LobbiesPlayed.String())
}

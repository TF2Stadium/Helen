package models_test

import (
	"testing"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	. "github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
}

func TestNewSub(t *testing.T) {
	testhelpers.CleanupDB()

	lobby := testhelpers.CreateLobby()
	lobby.Save()

	player := testhelpers.CreatePlayer()
	player.Save()

	tperr := lobby.AddPlayer(player, 0, "red", "scout1", "")
	assert.Nil(t, tperr)

	sub, err := NewSub(lobby.ID, player.SteamID)
	lobby.RemovePlayer(player)
	assert.Nil(t, err)

	db.DB.Save(sub)

	subs, err := GetPlayerSubs(player.SteamID)
	assert.Nil(t, err)
	assert.Equal(t, len(subs), 1)
	assert.Equal(t, subs[0].LobbyID, lobby.ID)

	player2 := testhelpers.CreatePlayer()
	player2.Save()
	tperr = lobby.AddPlayer(player2, 0, "red", "scout1", "")
	assert.Nil(t, tperr)

	err = db.DB.Where("lobby_id = ? AND steam_id = ?", lobby.ID, player.SteamID).First(sub).Error
	assert.Nil(t, err)
	assert.True(t, sub.Filled)
}

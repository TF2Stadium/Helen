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
	defer lobby.Close(false)
	lobby.Save()

	player := testhelpers.CreatePlayer()
	player.Save()

	tperr := lobby.AddPlayer(player, 0, "")
	assert.Nil(t, tperr)

	sub, err := NewSub(lobby.ID, player.ID)
	lobby.RemovePlayer(player)
	assert.Nil(t, err)

	db.DB.Save(sub)

	subs, err := GetPlayerSubs(player)
	assert.Nil(t, err)
	assert.Equal(t, len(subs), 1)
	assert.Equal(t, subs[0].LobbyID, lobby.ID)
	assert.Equal(t, subs[0].Team, "red")
	assert.Equal(t, subs[0].Class, "scout1")

	subs = GetAllSubs()
	assert.Equal(t, len(subs), 1)
	assert.Equal(t, subs[0].LobbyID, lobby.ID)
	assert.Equal(t, subs[0].Team, "red")
	assert.Equal(t, subs[0].Class, "scout1")

	player2 := testhelpers.CreatePlayer()
	player2.Save()
	tperr = lobby.AddPlayer(player2, 0, "")
	assert.Nil(t, tperr)

	err = db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(sub).Error
	assert.Nil(t, err)
	assert.True(t, sub.Filled)
}

package models_test

import (
	"testing"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/testhelpers"
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

	tperr := lobby.AddPlayer(player, 0, "red", "scout1")
	assert.Nil(t, tperr)

	sub, err := models.NewSub(lobby.ID, player.SteamId)
	assert.Nil(t, err)

	db.DB.Save(sub)

	subs, err := models.GetPlayerSubs(player.SteamId)
	assert.Nil(t, err)
	assert.Equal(t, len(subs), 1)
	assert.Equal(t, subs[0].MapName, lobby.MapName)
	assert.Equal(t, subs[0].LobbyID, lobby.ID)
}

package controllerhelpers

import (
	"testing"
	"time"

	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
}

func TestBanHelpers(t *testing.T) {
	migrations.TestCleanup()
	player, _ := models.NewPlayer("asdf")
	player.Save()

	assert.False(t, IsPlayerBanActive(player.BannedPlayUntil))
	assert.False(t, IsPlayerBanActive(player.BannedChatUntil))
	assert.False(t, IsPlayerBanActive(player.BannedCreateUntil))
	assert.False(t, IsPlayerBanActive(player.BannedFullUntil))

	future := time.Now().Add(time.Second * 10).Unix()

	tperr := SetPlayerBanTime(future, "play", player)
	assert.Nil(t, tperr)

	player2 := &models.Player{}
	err := database.DB.First(player2, player.ID).Error
	assert.Nil(t, err)

	assert.True(t, IsPlayerBanActive(player2.BannedPlayUntil))
	assert.False(t, IsPlayerBanActive(player2.BannedChatUntil))
	assert.False(t, IsPlayerBanActive(player2.BannedCreateUntil))
	assert.False(t, IsPlayerBanActive(player2.BannedFullUntil))

	tperr2 := SetPlayerBanTime(future, "create", player2)
	assert.Nil(t, tperr2)
	bans := GetPlayerBanTimes(player2)

	assert.Equal(t, future, bans["play"])
	assert.False(t, IsPlayerBanActive(bans["chat"]))
	assert.Equal(t, future, bans["create"])
	assert.False(t, IsPlayerBanActive(bans["full"]))

	tperr3 := UnbanPlayer("create", player2)
	assert.Nil(t, tperr3)
	assert.False(t, IsPlayerBanActive(player2.BannedCreateUntil))
}

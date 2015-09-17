// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"testing"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/testhelpers"
	"github.com/stretchr/testify/assert"
	"log"
	"time"
)

func init() {
	helpers.InitLogger()
}

func TestIsSpectating(t *testing.T) {
	testhelpers.CleanupDB()

	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "ugc", models.ServerRecord{}, 1)
	database.DB.Save(lobby)

	lobby2 := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "ugc", models.ServerRecord{}, 1)
	database.DB.Save(lobby2)

	player, _ := models.NewPlayer("asdf")
	database.DB.Save(player)

	isSpectating := player.IsSpectatingId(lobby.ID)
	assert.False(t, isSpectating)

	lobby.AddSpectator(player)

	isSpectating = player.IsSpectatingId(lobby.ID)
	assert.True(t, isSpectating)

	lobby2.AddSpectator(player)
	isSpectating2 := player.IsSpectatingId(lobby2.ID)
	assert.True(t, isSpectating2)

	specIds, specErr := player.GetSpectatingIds()
	assert.Nil(t, specErr)
	assert.Equal(t, []uint{lobby.ID, lobby2.ID}, specIds)

	lobby.RemoveSpectator(player)
	isSpectating = player.IsSpectatingId(lobby.ID)
	assert.False(t, isSpectating)
}

func TestPlayerInfoFetching(t *testing.T) {
	testhelpers.CleanupDB()

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
	testhelpers.CleanupDB()

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

func TestPlayerBanning(t *testing.T) {
	testhelpers.CleanupDB()
	player, _ := models.NewPlayer("76561197999073985")
	player.Save()

	assert.False(t, player.IsBanned(models.PlayerBanJoin))
	assert.False(t, player.IsBanned(models.PlayerBanCreate))
	assert.False(t, player.IsBanned(models.PlayerBanChat))
	assert.False(t, player.IsBanned(models.PlayerBanFull))

	past := time.Now().Add(time.Second * -10)
	player.BanUntil(past, models.PlayerBanJoin, "they suck")
	assert.False(t, player.IsBanned(models.PlayerBanJoin))

	future := time.Now().Add(time.Second * 10)
	player.BanUntil(future, models.PlayerBanJoin, "they suck")
	player.BanUntil(future, models.PlayerBanFull, "they suck")

	player2, _ := models.GetPlayerBySteamId(player.SteamId)
	assert.False(t, player2.IsBanned(models.PlayerBanCreate))
	assert.False(t, player2.IsBanned(models.PlayerBanChat))
	isBannedFull, untilFull := player2.IsBannedWithTime(models.PlayerBanFull)
	assert.True(t, isBannedFull)
	assert.True(t, future.Sub(untilFull) < time.Second)
	assert.True(t, untilFull.Sub(future) < time.Second)
	log.Println(future.Sub(untilFull))

	isBannedJoin, untilJoin := player2.IsBannedWithTime(models.PlayerBanJoin)
	assert.True(t, isBannedJoin)
	assert.True(t, future.Sub(untilJoin) < time.Second)
	assert.True(t, untilJoin.Sub(future) < time.Second)

	future2 := time.Now().Add(time.Second * 20)
	player2.BanUntil(future2, models.PlayerBanJoin, "they suck")
	isBannedJoin2, untilJoin2 := player2.IsBannedWithTime(models.PlayerBanJoin)
	assert.True(t, isBannedJoin2)
	assert.True(t, future2.Sub(untilJoin2) < time.Second)
	assert.True(t, untilJoin.Sub(future2) < time.Second)

	bans, err := player2.GetActiveBans()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(bans))

	player2.Unban(models.PlayerBanJoin)
	player2.Unban(models.PlayerBanFull)

	assert.False(t, player2.IsBanned(models.PlayerBanJoin))
	assert.False(t, player2.IsBanned(models.PlayerBanCreate))
	assert.False(t, player2.IsBanned(models.PlayerBanChat))
	assert.False(t, player2.IsBanned(models.PlayerBanFull))
}

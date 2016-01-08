// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"log"
	"testing"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	. "github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
	testhelpers.CleanupDB()
}

func TestGetPlayer(t *testing.T) {
	t.Parallel()
	player := testhelpers.CreatePlayer()
	player2, err := GetPlayerByID(player.ID)
	assert.NoError(t, err)
	assert.Equal(t, player.ID, player2.ID)
}

func TestIsSpectating(t *testing.T) {
	t.Parallel()

	lobby := testhelpers.CreateLobby()
	database.DB.Save(lobby)

	lobby2 := testhelpers.CreateLobby()
	database.DB.Save(lobby2)

	player := testhelpers.CreatePlayer()

	isSpectating := player.IsSpectatingID(lobby.ID)
	assert.False(t, isSpectating)

	lobby.AddSpectator(player)

	isSpectating = player.IsSpectatingID(lobby.ID)
	assert.True(t, isSpectating)

	lobby2.AddSpectator(player)
	isSpectating2 := player.IsSpectatingID(lobby2.ID)
	assert.True(t, isSpectating2)

	lobby.RemoveSpectator(player, false)
	isSpectating = player.IsSpectatingID(lobby.ID)
	assert.False(t, isSpectating)
}

func TestGetSpectatingIds(t *testing.T) {
	t.Parallel()

	player := testhelpers.CreatePlayer()

	specIds, specErr := player.GetSpectatingIds()
	assert.Nil(t, specErr)
	assert.Equal(t, len(specIds), 0)
	//assert.Equal(t, []uint{lobby.ID, lobby2.ID}, specIds)

	lobby1 := testhelpers.CreateLobby()
	database.DB.Save(lobby1)
	lobby1.AddSpectator(player)

	specIds, specErr = player.GetSpectatingIds()
	assert.Nil(t, specErr)
	assert.Equal(t, specIds[0], lobby1.ID)

	lobby2 := testhelpers.CreateLobby()
	database.DB.Save(lobby2)
	lobby2.AddSpectator(player)

	specIds, specErr = player.GetSpectatingIds()
	assert.Nil(t, specErr)
	for _, specID := range specIds {
		assert.True(t, lobby1.ID == specID || lobby2.ID == specID)
	}
}

func TestPlayerInfoFetching(t *testing.T) {
	t.Parallel()

	if config.Constants.SteamDevApiKey == "your steam dev api key" {
		return
	}

	// disable mock mode because we're actually testing it
	config.Constants.SteamApiMockUp = false

	player, playErr := NewPlayer("76561197999073985")
	assert.Nil(t, playErr)

	assert.Equal(t, "http://steamcommunity.com/id/nonagono/", player.Profileurl)
	assert.Regexp(t, "(.*)steamcommunity/public/images/avatars/(.*).jpg", player.Avatar)

	assert.True(t, player.GameHours >= 3000)

	player.Stats.PlayedCountIncrease(LobbyTypeSixes)
	player.Stats.PlayedCountIncrease(LobbyTypeHighlander)
	player.Stats.PlayedCountIncrease(LobbyTypeSixes) // sixes: 1 -> 2

	database.DB.Save(player)

	player2, err := GetPlayerWithStats(player.SteamID)
	assert.Nil(t, err)

	assert.Equal(t, 2, player2.Stats.PlayedSixesCount)
	assert.Equal(t, 1, player2.Stats.PlayedHighlanderCount)
	assert.Equal(t, "http://steamcommunity.com/id/nonagono/", player2.Profileurl)
}

func TestPlayerSettings(t *testing.T) {
	t.Parallel()

	player := testhelpers.CreatePlayer()

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
	t.Parallel()
	player := testhelpers.CreatePlayer()

	for ban := PlayerBanJoin; ban != PlayerBanFull; ban++ {
		assert.False(t, player.IsBanned(ban))
	}

	past := time.Now().Add(time.Second * -10)
	player.BanUntil(past, PlayerBanJoin, "they suck")
	assert.False(t, player.IsBanned(PlayerBanJoin))

	future := time.Now().Add(time.Second * 10)
	player.BanUntil(future, PlayerBanJoin, "they suck")
	player.BanUntil(future, PlayerBanFull, "they suck")

	player2, _ := GetPlayerBySteamID(player.SteamID)
	assert.False(t, player2.IsBanned(PlayerBanCreate))
	assert.False(t, player2.IsBanned(PlayerBanChat))
	isBannedFull, untilFull := player2.IsBannedWithTime(PlayerBanFull)
	assert.True(t, isBannedFull)
	assert.True(t, future.Sub(untilFull) < time.Second)
	assert.True(t, untilFull.Sub(future) < time.Second)
	log.Println(future.Sub(untilFull))

	isBannedJoin, untilJoin := player2.IsBannedWithTime(PlayerBanJoin)
	assert.True(t, isBannedJoin)
	assert.True(t, future.Sub(untilJoin) < time.Second)
	assert.True(t, untilJoin.Sub(future) < time.Second)

	future2 := time.Now().Add(time.Second * 20)
	player2.BanUntil(future2, PlayerBanJoin, "they suck")
	isBannedJoin2, untilJoin2 := player2.IsBannedWithTime(PlayerBanJoin)
	assert.True(t, isBannedJoin2)
	assert.True(t, future2.Sub(untilJoin2) < time.Second)
	assert.True(t, untilJoin.Sub(future2) < time.Second)

	bans, err := player2.GetActiveBans()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(bans))

	player2.Unban(PlayerBanJoin)
	player2.Unban(PlayerBanFull)

	for ban := PlayerBanJoin; ban != PlayerBanFull; ban++ {
		assert.False(t, player2.IsBanned(ban))
	}
}

func TestGetLobbyID(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	lobby.Save()

	player := testhelpers.CreatePlayer()
	player.Save()

	lobby.AddPlayer(player, 0, "")
	lobby.Save()

	id, err := player.GetLobbyID(false)
	assert.NoError(t, err)
	assert.Equal(t, id, lobby.ID)

	lobby.State = LobbyStateEnded
	lobby.Save()
	id, err = player.GetLobbyID(false)
	assert.Error(t, err)
	assert.Equal(t, id, uint(0))

	lobby.State = LobbyStateInProgress
	lobby.Save()

	//Exclude lobbies in progress
	id, err = player.GetLobbyID(true)
	assert.Error(t, err)
	assert.Equal(t, id, uint(0))

	//Include lobbies in progress
	id, err = player.GetLobbyID(false)
	assert.NoError(t, err)
	assert.Equal(t, id, lobby.ID)
}

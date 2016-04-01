// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package player_test

import (
	"testing"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/database"
	_ "github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	lobbypackage "github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/lobby/format"
	. "github.com/TF2Stadium/Helen/models/player"
	"github.com/stretchr/testify/assert"
)

func init() {
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

	if config.Constants.SteamDevAPIKey == "" {
		return
	}

	player, playErr := NewPlayer("76561197999073985")
	assert.Nil(t, playErr)

	assert.Equal(t, "http://steamcommunity.com/id/nonagono/", player.Profileurl)
	assert.Regexp(t, "(.*)steamcommunity/public/images/avatars/(.*).jpg", player.Avatar)

	assert.True(t, player.GameHours >= 3000)

	player.Stats.PlayedCountIncrease(format.Sixes)
	player.Stats.PlayedCountIncrease(format.Highlander)
	player.Stats.PlayedCountIncrease(format.Sixes) // sixes: 1 -> 2

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

	settings := player.Settings
	assert.Equal(t, 0, len(settings))

	player.SetSetting("foo", "bar")
	assert.Equal(t, player.GetSetting("foo"), "bar")

	player.SetSetting("hello", "world")
	assert.Equal(t, player.GetSetting("hello"), "world")
	assert.Len(t, player.Settings, 2)
}

func TestPlayerBanning(t *testing.T) {
	t.Parallel()
	player := testhelpers.CreatePlayer()

	for ban := BanJoin; ban != BanFull; ban++ {
		assert.False(t, player.IsBanned(ban))
	}

	past := time.Now().Add(time.Second * -10)
	player.BanUntil(past, BanJoin, "they suck", 0)
	assert.False(t, player.IsBanned(BanJoin))

	future := time.Now().Add(time.Second * 10)
	player.BanUntil(future, BanJoin, "they suck", 0)
	player.BanUntil(future, BanFull, "they suck", 0)

	player2, _ := GetPlayerBySteamID(player.SteamID)
	assert.True(t, player2.IsBanned(BanCreate))
	assert.True(t, player2.IsBanned(BanChat))
	isBannedFull, untilFull := player2.IsBannedWithTime(BanFull)
	assert.True(t, isBannedFull)
	assert.True(t, future.Sub(untilFull) < time.Second)
	assert.True(t, untilFull.Sub(future) < time.Second)

	isBannedJoin, untilJoin := player2.IsBannedWithTime(BanJoin)
	assert.True(t, isBannedJoin)
	assert.True(t, future.Sub(untilJoin) < time.Second)
	assert.True(t, untilJoin.Sub(future) < time.Second)

	future2 := time.Now().Add(time.Second * 20)
	player2.BanUntil(future2, BanJoin, "they suck", 0)

	bans, err := player2.GetActiveBans()
	assert.NoError(t, err)
	assert.Len(t, bans, 2)

	_, err = player2.GetActiveBan(BanJoin)
	assert.NoError(t, err)

	player2.Unban(BanJoin)
	player2.Unban(BanFull)

	for ban := BanJoin; ban != BanFull; ban++ {
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

	lobby.State = lobbypackage.Ended
	lobby.Save()
	id, err = player.GetLobbyID(false)
	assert.Error(t, err)
	assert.Equal(t, id, uint(0))

	lobby.State = lobbypackage.InProgress
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

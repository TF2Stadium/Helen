// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"math/rand"
	"testing"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	. "github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
	testhelpers.CleanupDB()
}

func TestDeleteUnusedServerRecords(t *testing.T) {
	var count int

	lobby := testhelpers.CreateLobby()
	lobby.Close(false)
	db.DB.Save(&ServerRecord{})

	DeleteUnusedServerRecords()

	err := db.DB.Table("server_records").Count(&count).Error
	assert.NoError(t, err)
	assert.Zero(t, count)
}

func TestLobbyCreation(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()

	lobby2, _ := GetLobbyByIDServer(lobby.ID)

	assert.Equal(t, lobby.ID, lobby2.ID)
	assert.Equal(t, lobby.ServerInfo.Host, lobby2.ServerInfo.Host)
	assert.Equal(t, lobby.ServerInfo.ID, lobby2.ServerInfo.ID)

	lobby.MapName = "cp_granary"
	lobby.Save()

	db.DB.First(lobby2)
	assert.Equal(t, "cp_granary", lobby2.MapName)
}

func TestLobbyClose(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	lobby.Save()

	req := &Requirement{
		LobbyID: lobby.ID,
	}
	req.Save()
	lobby.Close(true)
	var count int
	db.DB.Table("requirements").Where("lobby_id = ?", lobby.ID).Count(&count)
	assert.Zero(t, count)

	db.DB.Table("server_records").Where("id = ?", lobby.ServerInfoID).Count(&count)
	assert.Zero(t, count)
	lobby, _ = GetLobbyByID(lobby.ID)
	assert.Equal(t, lobby.State, LobbyStateEnded)
}

func TestLobbyAdd(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()

	var players []*Player

	for i := 0; i < 12; i++ {
		player := testhelpers.CreatePlayer()
		players = append(players, player)
	}

	// add player
	err := lobby.AddPlayer(players[0], 0, "")
	assert.Nil(t, err)

	slot, err2 := lobby.GetPlayerSlot(players[0])
	assert.Zero(t, slot)
	assert.Nil(t, err2)

	id, err3 := lobby.GetPlayerIDBySlot(0)
	assert.Equal(t, id, players[0].ID)
	assert.Nil(t, err3)

	// try to switch slots
	err = lobby.AddPlayer(players[0], 1, "")
	assert.Nil(t, err)

	slot, err2 = lobby.GetPlayerSlot(players[0])
	assert.Equal(t, slot, 1)
	assert.Nil(t, err2)

	// this should be empty now
	id, err3 = lobby.GetPlayerIDBySlot(0)
	assert.NotNil(t, err3)

	// try to add a second player to the same slot
	err = lobby.AddPlayer(players[1], 1, "")
	assert.NotNil(t, err)

	// try to add a player to a wrong slot slot
	err = lobby.AddPlayer(players[2], 55, "")
	assert.NotNil(t, err)

	lobby2 := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby2.Save()

	// try to add a player while they're in another lobby
	lobby.State = LobbyStateInProgress
	lobby.Save()
	err = lobby2.AddPlayer(players[0], 1, "")
	assert.Nil(t, err)

	var count int
	db.DB.Table("substitutes").Where("lobby_id = ?", lobby.ID).Count(&count)
	assert.Equal(t, count, 1)
}

func TestLobbyRemove(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()

	player := testhelpers.CreatePlayer()

	// add player
	err := lobby.AddPlayer(player, 0, "")
	assert.Nil(t, err)

	// remove player
	err = lobby.RemovePlayer(player)
	assert.Nil(t, err)

	// this should be empty now
	_, err2 := lobby.GetPlayerIDBySlot(0)
	assert.NotNil(t, err2)

	// can add player again
	err = lobby.AddPlayer(player, 0, "")
	assert.Nil(t, err)
}

func TestLobbyBan(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()

	player := testhelpers.CreatePlayer()

	// add player
	err := lobby.AddPlayer(player, 0, "")
	assert.Nil(t, err)

	// ban player
	err = lobby.RemovePlayer(player)
	lobby.BanPlayer(player)
	assert.Nil(t, err)

	// should not be able to add again
	err = lobby.AddPlayer(player, 5, "")
	assert.NotNil(t, err)
}

func TestReadyPlayer(t *testing.T) {
	t.Parallel()
	player := testhelpers.CreatePlayer()

	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()
	lobby.AddPlayer(player, 0, "")

	lobby.ReadyPlayer(player)
	ready, err := lobby.IsPlayerReady(player)
	assert.Equal(t, ready, true)
	assert.Nil(t, err)

	lobby.UnreadyPlayer(player)
	lobby.ReadyPlayer(player)
	ready, err = lobby.IsPlayerReady(player)
	assert.Equal(t, ready, true)
	assert.Nil(t, err)
}

func TestSetInGame(t *testing.T) {
	t.Parallel()
	player := testhelpers.CreatePlayer()

	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()
	lobby.AddPlayer(player, 0, "")
	lobby.SetInGame(player)

	slot, err := lobby.GetPlayerSlotObj(player)
	assert.Nil(t, err)
	assert.True(t, slot.InGame)
}

func TestSetNotInGame(t *testing.T) {
	t.Parallel()
	player := testhelpers.CreatePlayer()

	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()
	lobby.AddPlayer(player, 0, "")
	lobby.SetInGame(player)
	lobby.SetNotInGame(player)

	slot, err := lobby.GetPlayerSlotObj(player)
	assert.Nil(t, err)
	assert.False(t, slot.InGame)
}
func TestIsEveryoneReady(t *testing.T) {
	t.Parallel()
	player := testhelpers.CreatePlayer()

	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()
	lobby.AddPlayer(player, 0, "")
	lobby.ReadyPlayer(player)
	assert.Equal(t, lobby.IsEveryoneReady(), false)

	for i := 1; i < 12; i++ {
		player := testhelpers.CreatePlayer()
		lobby.AddPlayer(player, i, "")
		lobby.ReadyPlayer(player)
	}
	assert.Equal(t, lobby.IsEveryoneReady(), true)
}

func TestUnreadyPlayer(t *testing.T) {
	t.Parallel()
	player := testhelpers.CreatePlayer()

	player.Save()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()
	lobby.AddPlayer(player, 0, "")

	lobby.ReadyPlayer(player)
	lobby.UnreadyPlayer(player)
	ready, err := lobby.IsPlayerReady(player)
	assert.Equal(t, ready, false)
	assert.Nil(t, err)
}

func TestSpectators(t *testing.T) {
	t.Parallel()

	player := testhelpers.CreatePlayer()

	player2 := testhelpers.CreatePlayer()

	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()

	err := lobby.AddSpectator(player)
	assert.Nil(t, err)

	var specs []Player
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 1, len(specs))

	err = lobby.AddSpectator(player2)
	assert.Nil(t, err)

	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 2, len(specs))
	assert.Equal(t, true, specs[0].IsSpectatingID(lobby.ID))

	err = lobby.RemoveSpectator(player, false)
	assert.Nil(t, err)

	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 1, len(specs))

	// adding the same player again should not increase the count
	err = lobby.AddSpectator(player2)
	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 1, len(specs))

	// adding a player should remove them from spectators
	lobby.AddPlayer(player2, 11, "")
	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Zero(t, len(specs))
}

func TestUnreadyAllPlayers(t *testing.T) {
	t.Parallel()

	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()

	for i := 0; i < 12; i++ {
		player := testhelpers.CreatePlayer()
		lobby.AddPlayer(player, i, "")
		lobby.ReadyPlayer(player)
	}

	err := lobby.UnreadyAllPlayers()
	assert.Nil(t, err)
	ready := lobby.IsEveryoneReady()
	assert.Equal(t, ready, false)
}

func TestRemoveUnreadyPlayers(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	lobby.Save()

	var players []*Player
	for i := 0; i < 12; i++ {
		player := testhelpers.CreatePlayer()

		lobby.AddPlayer(player, i, "")
		players = append(players, player)
	}

	err := lobby.RemoveUnreadyPlayers(true)
	assert.Nil(t, err)

	for i := 0; i < 12; i++ {
		var count int
		_, err := lobby.GetPlayerIDBySlot(i)
		assert.Error(t, err)

		db.DB.Table("spectators_players_lobbies").Where("lobby_id = ? AND player_id = ?", lobby.ID, players[i].ID).Count(&count)
		assert.Equal(t, count, 1)
	}
}

func TestUpdateStats(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	var players []*Player

	for i := 0; i < 6; i++ {
		players = append(players, testhelpers.CreatePlayer())
	}
	for i, player := range players {
		err := lobby.AddPlayer(player, i, "")
		assert.NoError(t, err)
	}

	lobby.UpdateStats()
	for _, player := range players {
		db.DB.Preload("Stats").First(player, player.ID)
		assert.Equal(t, player.Stats.PlayedSixesCount, 1)
	}
}

func TestNotInGameSub(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	var players []*Player
	var ingame int

	for i := 0; i < 12; i++ {
		players = append(players, testhelpers.CreatePlayer())
	}
	for i, player := range players {
		err := lobby.AddPlayer(player, i, "")
		assert.NoError(t, err)
		if rand.Intn(2) == 1 {
			ingame++
			lobby.SetInGame(player)
		}
	}

	lobby.SubNotInGamePlayers()
	assert.Equal(t, lobby.GetPlayerNumber(), ingame)

	var subcount int
	db.DB.Table("substitutes").Where("lobby_id = ?", lobby.ID).Count(&subcount)
	assert.Equal(t, subcount, 12-ingame)
}

func TestSlotRequirements(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	player := testhelpers.CreatePlayer()
	req := &Requirement{
		LobbyID: lobby.ID,
		Slot:    0,
		Hours:   1,
		Lobbies: 1,
	}
	req.Save()

	assert.True(t, lobby.HasRequirements(0))
	err := lobby.AddPlayer(player, 0, "")
	assert.Equal(t, err, ReqHoursErr)

	player.GameHours = 2
	player.Save()

	err = lobby.AddPlayer(player, 0, "")
	assert.Equal(t, err, ReqLobbiesErr)

	player, _ = GetPlayerWithStats(player.SteamID)
	player.Stats.PlayedCountIncrease(lobby.Type)

	err = lobby.AddPlayer(player, 0, "")
	assert.NoError(t, err)

	//Adding a player to another slot shouldn't return any errors
	// req = &Requirement{
	// 	LobbyID: lobby.ID,
	// 	Slot:    -1,
	// 	Hours:   1,
	// 	Lobbies: 1,
	// }
	player2 := testhelpers.CreatePlayer()
	err = lobby.AddPlayer(player2, 1, "")
	assert.NoError(t, err)
}

func TestGeneralRequirements(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false)
	player := testhelpers.CreatePlayer()
	req := &Requirement{
		LobbyID: lobby.ID,
		Slot:    -1,
		Hours:   1,
		Lobbies: 1,
	}
	req.Save()

	err := lobby.AddPlayer(player, 0, "")
	assert.Equal(t, err, ReqHoursErr)

	err = lobby.AddPlayer(player, 3, "")
	assert.Equal(t, err, ReqHoursErr)
}

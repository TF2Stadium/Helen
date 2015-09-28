// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"fmt"
	"strconv"
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

func TestLobbyCreation(t *testing.T) {
	testhelpers.CleanupDB()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "ugc", models.ServerRecord{0, "testip", "", ""}, 0)
	lobby.Save()

	lobby2, _ := models.GetLobbyById(lobby.ID)

	assert.Equal(t, lobby.ID, lobby2.ID)
	assert.Equal(t, lobby.ServerInfo.Host, lobby2.ServerInfo.Host)
	assert.Equal(t, lobby.ServerInfo.ID, lobby2.ServerInfo.ID)

	lobby.MapName = "cp_granary"
	lobby.Save()

	db.DB.First(lobby2)
	assert.Equal(t, "cp_granary", lobby2.MapName)
}

func TestLobbyAdd(t *testing.T) {
	testhelpers.CleanupDB()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "ugc", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()

	var players []*models.Player

	for i := 0; i < 12; i++ {
		player, playErr := models.NewPlayer("p" + fmt.Sprint(i))
		assert.Nil(t, playErr)

		player.Save()
		players = append(players, player)
	}

	// add player
	err := lobby.AddPlayer(players[0], 0)
	assert.Nil(t, err)

	slot, err2 := lobby.GetPlayerSlot(players[0])
	assert.Equal(t, slot, 0)
	assert.Nil(t, err2)

	id, err3 := lobby.GetPlayerIdBySlot(0)
	assert.Equal(t, id, players[0].ID)
	assert.Nil(t, err3)

	// try to switch slots
	err = lobby.AddPlayer(players[0], 1)
	assert.Nil(t, err)

	slot, err2 = lobby.GetPlayerSlot(players[0])
	assert.Equal(t, slot, 1)
	assert.Nil(t, err2)

	// this should be empty now
	id, err3 = lobby.GetPlayerIdBySlot(0)
	assert.NotNil(t, err3)

	// try to add a second player to the same slot
	err = lobby.AddPlayer(players[1], 1)
	assert.NotNil(t, err)

	// try to add a player to a wrong slot slot
	err = lobby.AddPlayer(players[2], 55)
	assert.NotNil(t, err)

	lobby2 := models.NewLobby("cp_granary", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby2.Save()

	// try to add a player while they're in another lobby
	err = lobby.AddPlayer(players[0], 1)
	assert.NotNil(t, err)
}

func TestLobbyRemove(t *testing.T) {
	testhelpers.CleanupDB()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()

	player, playErr := models.NewPlayer("1235")
	assert.Nil(t, playErr)
	player.Save()

	// add player
	err := lobby.AddPlayer(player, 0)
	assert.Nil(t, err)

	// remove player
	err = lobby.RemovePlayer(player)
	assert.Nil(t, err)

	// this should be empty now
	_, err2 := lobby.GetPlayerIdBySlot(0)
	assert.NotNil(t, err2)

	// can add player again
	err = lobby.AddPlayer(player, 0)
	assert.Nil(t, err)
}

func TestLobbyBan(t *testing.T) {
	testhelpers.CleanupDB()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()

	player, playErr := models.NewPlayer("1235")
	assert.Nil(t, playErr)
	player.Save()

	// add player
	err := lobby.AddPlayer(player, 0)
	assert.Nil(t, err)

	// ban player
	err = lobby.RemovePlayer(player)
	lobby.BanPlayer(player)
	assert.Nil(t, err)

	// should not be able to add again
	err = lobby.AddPlayer(player, 5)
	assert.NotNil(t, err)
}

func TestReadyPlayer(t *testing.T) {
	testhelpers.CleanupDB()
	player, playErr := models.NewPlayer("testing")
	assert.Nil(t, playErr)

	player.Save()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()
	lobby.AddPlayer(player, 0)

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

func TestIsEveryoneReady(t *testing.T) {
	testhelpers.CleanupDB()
	player, playErr := models.NewPlayer("0")
	assert.Nil(t, playErr)

	player.Save()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()
	lobby.AddPlayer(player, 0)
	lobby.ReadyPlayer(player)
	assert.Equal(t, lobby.IsEveryoneReady(), false)

	for i := 1; i < 12; i++ {
		player, playErr = models.NewPlayer(strconv.Itoa(i))
		assert.Nil(t, playErr)
		player.Save()
		lobby.AddPlayer(player, i)
		lobby.ReadyPlayer(player)
	}
	assert.Equal(t, lobby.IsEveryoneReady(), true)
}

func TestUnreadyPlayer(t *testing.T) {
	testhelpers.CleanupDB()
	player, playErr := models.NewPlayer("testing")
	assert.Nil(t, playErr)

	player.Save()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()
	lobby.AddPlayer(player, 0)

	lobby.ReadyPlayer(player)
	lobby.UnreadyPlayer(player)
	ready, err := lobby.IsPlayerReady(player)
	assert.Equal(t, ready, false)
	assert.Nil(t, err)
}

func TestSpectators(t *testing.T) {
	testhelpers.CleanupDB()

	player, playErr := models.NewPlayer("apple")
	assert.Nil(t, playErr)

	player.Save()

	player2, playErr2 := models.NewPlayer("testing1")
	assert.Nil(t, playErr2)
	player2.Save()

	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()

	err := lobby.AddSpectator(player)
	assert.Nil(t, err)

	var specs []models.Player
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 1, len(specs))

	err = lobby.AddSpectator(player2)
	assert.Nil(t, err)

	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 2, len(specs))
	assert.Equal(t, true, specs[0].IsSpectatingId(lobby.ID))

	err = lobby.RemoveSpectator(player)
	assert.Nil(t, err)

	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 1, len(specs))

	// adding the same player again should not increase the count
	err = lobby.AddSpectator(player2)
	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 1, len(specs))

	// players in lobby should be removed from it if added as spectator
	lobby.AddPlayer(player, 10)
	err = lobby.AddSpectator(player)
	assert.Nil(t, err)
	_, terr := lobby.GetPlayerSlot(player)
	assert.NotNil(t, terr)

	// adding a player should remove them from spectators
	lobby.AddPlayer(player2, 11)
	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 0, len(specs))
}

func TestUnreadyAllPlayers(t *testing.T) {
	testhelpers.CleanupDB()

	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()

	for i := 0; i < 12; i++ {
		player, playErr := models.NewPlayer(strconv.Itoa(i))
		assert.Nil(t, playErr)
		player.Save()
		lobby.AddPlayer(player, i)
		lobby.ReadyPlayer(player)
	}

	err := lobby.UnreadyAllPlayers()
	assert.Nil(t, err)
	ready := lobby.IsEveryoneReady()
	assert.Equal(t, ready, false)
}

func TestRemoveUnreadyPlayers(t *testing.T) {
	testhelpers.CleanupDB()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, "", models.ServerRecord{0, "", "", ""}, 0)
	lobby.Save()

	for i := 0; i < 12; i++ {
		player, playErr := models.NewPlayer(strconv.Itoa(i))
		assert.Nil(t, playErr)
		player.Save()
		lobby.AddPlayer(player, i)
	}

	err := lobby.RemoveUnreadyPlayers()
	assert.Nil(t, err)

	for i := 0; i < 12; i++ {
		_, err := lobby.GetPlayerIdBySlot(i)
		assert.NotNil(t, err)
	}
}

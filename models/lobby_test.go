package models_test

import (
	"fmt"
	"testing"

	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/database/migrations"
	"github.com/TF2Stadium/Server/models"
	"github.com/stretchr/testify/assert"
)

func TestLobbyCreation(t *testing.T) {
	migrations.TestCleanup()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, 0)
	lobby.Save()

	lobby2 := &models.Lobby{}
	db.DB.First(lobby2)

	assert.Equal(t, lobby.ID, lobby2.ID)

	lobby.MapName = "cp_granary"
	lobby.Save()

	db.DB.First(lobby2)
	assert.Equal(t, "cp_granary", lobby2.MapName)
}

func TestLobbyAdd(t *testing.T) {
	migrations.TestCleanup()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, 0)
	lobby.Save()

	var players []*models.Player

	for i := 0; i < 12; i++ {
		player := models.NewPlayer("p" + fmt.Sprint(i))
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

	lobby2 := models.NewLobby("cp_granary", models.LobbyTypeSixes, 0)
	lobby2.Save()

	// try to add a player while they're in another lobby
	err = lobby.AddPlayer(players[0], 1)
	assert.NotNil(t, err)
}

func TestLobbyRemove(t *testing.T) {
	migrations.TestCleanup()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, 0)
	lobby.Save()

	player := models.NewPlayer("1235")
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
	migrations.TestCleanup()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, 0)
	lobby.Save()

	player := models.NewPlayer("1235")
	player.Save()

	// add player
	err := lobby.AddPlayer(player, 0)
	assert.Nil(t, err)

	// ban player
	err = lobby.KickAndBanPlayer(player)
	assert.Nil(t, err)

	// should not be able to add again
	err = lobby.AddPlayer(player, 5)
	assert.NotNil(t, err)
}

func TestReadyPlayer(t *testing.T) {
	migrations.TestCleanup()
	player := models.NewPlayer("testing")
	player.Save()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, 0)
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

func TestUnreadyPlayer(t *testing.T) {
	migrations.TestCleanup()
	player := models.NewPlayer("testing")
	player.Save()
	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, 0)
	lobby.Save()
	lobby.AddPlayer(player, 0)

	lobby.ReadyPlayer(player)
	lobby.UnreadyPlayer(player)
	ready, err := lobby.IsPlayerReady(player)
	assert.Equal(t, ready, false)
	assert.Nil(t, err)
}

func TestSpectators(t *testing.T) {
	migrations.TestCleanup()
	player := models.NewPlayer("testing")
	player.Save()
	player2 := models.NewPlayer("testing1")
	player2.Save()

	lobby := models.NewLobby("cp_badlands", models.LobbyTypeSixes, 0)
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

	// players in lobby should not be added as spectators
	lobby.AddPlayer(player, 10)
	err = lobby.AddSpectator(player)
	assert.NotNil(t, err)

	// adding a player should remove them from spectators
	lobby.AddPlayer(player2, 11)
	specs = nil
	db.DB.Model(lobby).Association("Spectators").Find(&specs)
	assert.Equal(t, 0, len(specs))
}

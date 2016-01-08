package rpc_test

import (
	"testing"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/rpc"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
	testhelpers.CleanupDB()
}

func TestPlayerDisc(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	player := testhelpers.CreatePlayer()

	lobby.AddPlayer(player, 0, "")
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("in_game", true)

	e := rpc.Event{
		Name:     rpc.PlayerDisconnected,
		LobbyID:  lobby.ID,
		PlayerID: player.ID,
	}
	e.Handle(e, &struct{}{})

	var count int
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ? AND in_game = ?", lobby.ID, player.ID, false).Count(&count)
	assert.Equal(t, count, 1)
}

func TestPlayerConn(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()
	player := testhelpers.CreatePlayer()

	lobby.AddPlayer(player, 0, "")
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("in_game", false)
	e := rpc.Event{
		Name:     rpc.PlayerConnected,
		LobbyID:  lobby.ID,
		PlayerID: player.ID,
	}
	e.Handle(e, &struct{}{})

	var count int
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ? AND in_game = ?", lobby.ID, player.ID, true).Count(&count)
	assert.Equal(t, count, 1)
}

func TestDisconnectedFromServer(t *testing.T) {
	t.Parallel()
	lobby := testhelpers.CreateLobby()

	e := rpc.Event{
		Name:    rpc.DisconnectedFromServer,
		LobbyID: lobby.ID,
	}

	e.Handle(e, &struct{}{})
	assert.Equal(t, lobby.CurrentState(), models.LobbyStateEnded)
}

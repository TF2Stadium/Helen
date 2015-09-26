// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"fmt"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/testhelpers"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotAuthedLobbyData(t *testing.T) {
	testhelpers.CleanupDB()

	lobby := testhelpers.CreateLobby()
	lobby.State = models.LobbyStateWaiting
	lobby.Save()

	pSocket1 := testhelpers.NewFakeSocket()

	SocketInit(pSocket1)
	msg1 := pSocket1.GetNextNamedMessage("lobbyListData")
	assert.NotNil(t, msg1)
	//	msg1.
	assert.Equal(t, int(lobby.ID), msg1.Get("lobbies").GetIndex(0).Get("id").MustInt())

	msg2 := pSocket1.GetNextNamedMessage("lobbyData")
	assert.Nil(t, msg2)
}

func TestAuthedLobbyData(t *testing.T) {
	testhelpers.CleanupDB()

	lobby := testhelpers.CreateLobby()
	lobby.State = models.LobbyStateWaiting
	lobby.Save()

	lobby2 := testhelpers.CreateLobby()
	lobby2.State = models.LobbyStateWaiting
	lobby2.Save()

	player1 := testhelpers.CreatePlayer()
	pSocket1 := testhelpers.NewFakeSocket()
	pSocket1.FakeAuthenticate(player1)

	SocketInit(pSocket1)

	res1, _ := pSocket1.SimRequest("lobbyJoin", fmt.Sprintf(`{"id": %d, "team": "red", "class": "scout1"}`, lobby.ID))
	res2, _ := pSocket1.SimRequest("lobbySpectatorJoin", fmt.Sprintf(`{"id": %d}`, lobby2.ID))
	testhelpers.UnpackSuccessResponse(t, res1)
	testhelpers.UnpackSuccessResponse(t, res2)

	msg1 := pSocket1.GetNextNamedMessage("lobbyListData")
	assert.NotNil(t, msg1)
	assert.Equal(t, int(lobby2.ID), msg1.Get("lobbies").GetIndex(0).Get("id").MustInt())

	msg2 := pSocket1.GetNextNamedMessage("lobbyData")
	msg3 := pSocket1.GetNextNamedMessage("lobbyData")
	assert.NotNil(t, msg2)
	assert.NotNil(t, msg3)
	ids := []uint{uint(msg2.Get("id").MustInt()), uint(msg3.Get("id").MustInt())}
	assert.True(t, ids[0] == lobby.ID || ids[1] == lobby.ID)
	assert.True(t, ids[0] == lobby2.ID || ids[1] == lobby2.ID)
}

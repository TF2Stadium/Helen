// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"github.com/TF2Stadium/Helen/testhelpers"
	"testing"
	//	"github.com/TF2Stadium/Helen/helpers"
	//	"github.com/stretchr/testify/assert"
	"fmt"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/stretchr/testify/assert"
	"os"
)

func TestMain(m *testing.M) {
	testhelpers.SetupFakeSockets()
	helpers.InitAuthorization()
	database.Init()
	res := m.Run()
	database.DB.Close()
	os.Exit(res)
}

func TestChatSend(t *testing.T) {
	testhelpers.CleanupDB()

	player1 := testhelpers.CreatePlayer()
	pSocket1 := testhelpers.NewFakeSocket()
	pSocket1.FakeAuthenticate(player1)

	player2 := testhelpers.CreatePlayer()
	pSocket2 := testhelpers.NewFakeSocket()
	pSocket2.FakeAuthenticate(player2)

	player3 := testhelpers.CreatePlayer()
	pSocket3 := testhelpers.NewFakeSocket()
	pSocket3.FakeAuthenticate(player3)

	SocketInit(pSocket1)
	SocketInit(pSocket2)
	SocketInit(pSocket3)

	// can send chat messages to global
	res1, _ := pSocket1.SimRequest("chatSend", `{"room": 0, "message": "o hai"}`)
	testhelpers.UnpackSuccessResponse(t, res1)

	msg1 := pSocket1.GetNextNamedMessage("chatReceive")
	assert.NotNil(t, msg1)
	assert.Equal(t, 0, msg1.Get("room").MustInt())
	assert.Equal(t, "o hai", msg1.Get("message").MustString())
	assert.Equal(t, player1.SteamId, msg1.Get("player").Get("steamid").MustString())

	msg2 := pSocket2.GetNextNamedMessage("chatReceive")
	assert.NotNil(t, msg2)
	assert.Equal(t, 0, msg2.Get("room").MustInt())
	assert.Equal(t, "o hai", msg2.Get("message").MustString())
	assert.Equal(t, player1.SteamId, msg2.Get("player").Get("steamid").MustString())

	// just constume the message for future tests
	pSocket3.GetNextNamedMessage("chatReceive")

	// can't send chat messages to lobbies they're not in
	res2, _ := pSocket1.SimRequest("chatSend", `{"room": 5555, "message": "o hai"}`)
	testhelpers.UnpackFailureResponse(t, res2)

	// can send chat messages to lobbies they are in
	lobby := testhelpers.CreateLobby()
	pSocket1.SimRequest("lobbyJoin", fmt.Sprintf(`{"id": %d, "team": "red", "class": "scout1"}`, lobby.ID))
	pSocket2.SimRequest("lobbyJoin", fmt.Sprintf(`{"id": %d, "team": "blu", "class": "scout1"}`, lobby.ID))

	res3, _ := pSocket2.SimRequest("chatSend", fmt.Sprintf(`{"room":%d, "message": "o hai"}`, lobby.ID))
	testhelpers.UnpackSuccessResponse(t, res3)

	msg3 := pSocket1.GetNextNamedMessage("chatReceive")
	assert.NotNil(t, msg3)
	assert.Equal(t, int(lobby.ID), msg3.Get("room").MustInt())
	assert.Equal(t, "o hai", msg3.Get("message").MustString())
	assert.Equal(t, player2.SteamId, msg3.Get("player").Get("steamid").MustString())

	msg4 := pSocket2.GetNextNamedMessage("chatReceive")
	assert.NotNil(t, msg4)
	assert.Equal(t, int(lobby.ID), msg4.Get("room").MustInt())
	assert.Equal(t, "o hai", msg4.Get("message").MustString())
	assert.Equal(t, player2.SteamId, msg4.Get("player").Get("steamid").MustString())

	// player not in lobby shouldn't receive a message
	msg5 := pSocket3.GetNextNamedMessage("chatReceive")
	assert.Nil(t, msg5)
}

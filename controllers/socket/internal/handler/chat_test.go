package handler_test

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/TF2Stadium/Helen/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestChatSend(t *testing.T) {
	testhelpers.CleanupDB()
	server := testhelpers.StartServer(testhelpers.NewSockets())
	defer server.Close()

	steamid := strconv.Itoa(rand.Int())
	client := testhelpers.NewClient()
	testhelpers.Login(steamid, client)
	conn, err := testhelpers.ConnectWS(client)
	defer conn.Close()
	assert.NoError(t, err)
	_, err = testhelpers.ReadMessages(conn, testhelpers.InitMessages, nil)
	assert.NoError(t, err)

	conn.WriteJSON(
		map[string]interface{}{
			"id": "1",
			"data": map[string]interface{}{
				"request": "chatSend",
				"message": "testerino",
				"room":    0,
			},
		})

	messages, err := testhelpers.ReadMessages(conn, 2, nil)
	for _, message := range messages {
		_, ok := message["success"]
		if ok {
			assert.True(t, message["success"].(bool))
		} else {
			assert.Equal(t, message["request"], "chatReceive")
			assert.Equal(t, message["data"].(map[string]interface{})["player"].(map[string]interface{})["steamid"], steamid)
		}
	}

	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request":             "lobbyCreate",
			"map":                 "cp_badlands",
			"type":                "6s",
			"league":              "etf2l",
			"server":              "testerino",
			"rconpwd":             "testerino",
			"whitelistID":         123,
			"mumbleRequired":      true,
			"password":            nil,
			"steamGroupWhitelist": nil,
		}}

	conn.WriteJSON(args)
	testhelpers.ReadMessages(conn, 2, nil)

	testhelpers.SocketJoinLobby(conn)

	time.Sleep(1 * time.Second)
	conn.WriteJSON(
		map[string]interface{}{
			"id": "1",
			"data": map[string]interface{}{
				"request": "chatSend",
				"message": "testerino",
				"room":    1,
			},
		})
	messages, _ = testhelpers.ReadMessages(conn, 2, nil)
	for _, message := range messages {
		_, ok := message["success"]
		if ok {
			assert.True(t, message["success"].(bool))
		} else {
			assert.Equal(t, message["request"], "chatReceive")
			assert.Equal(t, message["data"].(map[string]interface{})["player"].(map[string]interface{})["steamid"], steamid)
		}
	}
}

package handler_test

import (
	"math/rand"
	"strconv"
	"testing"

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

}

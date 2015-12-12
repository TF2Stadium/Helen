package controllers_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	//"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestSocketInfo(t *testing.T) {
	server := testhelpers.StartServer(testhelpers.NewSockets())
	defer server.Close()

	client := testhelpers.NewClient()
	testhelpers.Login(strconv.Itoa(rand.Int()), client)
	conn, err := testhelpers.ConnectWS(client)
	defer conn.Close()
	assert.NoError(t, err)

	_, err = testhelpers.ReadMessages(conn, testhelpers.InitMessages, nil)
	assert.NoError(t, err)

	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "getSocketInfo",
		},
	}

	reply, err := testhelpers.EmitJSONWithReply(conn, args)
	assert.NoError(t, err)
	data := reply["data"].(map[string]interface{})
	assert.Equal(t, data["rooms"].([]interface{})[0].(string), "0_public")
	//t.Logf("%v", reply)

}

func TestLobbyCreate(t *testing.T) {
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

	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request":        "lobbyCreate",
			"map":            "cp_badlands",
			"type":           "6s",
			"league":         "etf2l",
			"server":         "testerino",
			"rconpwd":        "testerino",
			"whitelistID":    123,
			"mumbleRequired": true,
		}}

	conn.WriteJSON(args)

	messages, err := testhelpers.ReadMessages(conn, 2, nil)
	for _, message := range messages {
		_, ok := message["success"]
		if ok {
			assert.True(t, message["success"].(bool))
		} else {
			assert.Equal(t, message["request"], "lobbyListData")
		}
	}

	lobby, err := models.GetLobbyById(1)
	assert.NoError(t, err)
	assert.Equal(t, lobby.CreatedBySteamID, steamid)
}

type newLobby struct {
	args map[string]interface{}
	i    int
}

func BenchmarkLobbyCreate(b *testing.B) {
	server := testhelpers.StartServer(testhelpers.NewSockets())
	defer server.Close()

	steamid := strconv.Itoa(rand.Int())
	client := testhelpers.NewClient()
	testhelpers.Login(steamid, client)
	conn, err := testhelpers.ConnectWS(client)
	assert.NoError(b, err)

	_, err = testhelpers.ReadMessages(conn, testhelpers.InitMessages, nil)
	assert.NoError(b, err)

	argsChan := make(chan newLobby)
	wg := new(sync.WaitGroup)
	go func() {
		for {
			args := <-argsChan
			//helpers.Logger.Debug("%d", args.i)
			err := conn.WriteJSON(args.args)
			if err != nil {
				b.Error(err)
				b.FailNow()
			}

			messages, _ := testhelpers.ReadMessages(conn, 2, nil)

			//helpers.Logger.Debug("%v", messages)
			for _, message := range messages {
				_, ok := message["success"]
				if ok {
					assert.True(b, message["success"].(bool))
					break
				}
			}
			wg.Done()
		}
	}()

	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args := map[string]interface{}{
			"id": "1",
			"data": map[string]interface{}{
				"request":        "lobbyCreate",
				"map":            "cp_badlands",
				"type":           "6s",
				"league":         "etf2l",
				"server":         strconv.Itoa(i + rand.Int()),
				"rconpwd":        "testerino",
				"whitelistID":    123,
				"mumbleRequired": true,
			}}
		argsChan <- newLobby{args, i}

	}

	wg.Wait()
}

func TestLobbyJoin(t *testing.T) {
	server := testhelpers.StartServer(testhelpers.NewSockets())
	defer server.Close()

	steamid := strconv.Itoa(rand.Int())
	client := testhelpers.NewClient()
	testhelpers.Login(steamid, client)
	player, tperr := models.GetPlayerBySteamId(steamid)
	assert.NoError(t, tperr)
	conn, err := testhelpers.ConnectWS(client)
	defer conn.Close()
	assert.NoError(t, err)
	_, err = testhelpers.ReadMessages(conn, testhelpers.InitMessages, nil)
	assert.NoError(t, err)

	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbyJoin",
			"id":      1,
			"class":   "scout1",
			"team":    "red",
		}}

	conn.WriteJSON(args)

	messages, err := testhelpers.ReadMessages(conn, 2, nil)
	for _, message := range messages {
		_, ok := message["success"]
		if ok {
			assert.True(t, message["success"].(bool))
		} else {
			assert.Equal(t, message["request"], "lobbyJoined")
		}
	}

	id, tperr := player.GetLobbyId()
	assert.NoError(t, tperr)
	if id != 1 {
		t.Fatal("Got wrong ID")
	}

	lobby, tperr := models.GetLobbyById(1)
	assert.NoError(t, tperr)
	assert.Equal(t, lobby.GetPlayerNumber(), 1)
}

func TestSpectatorJoin(t *testing.T) {
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
				"request": "lobbySpectatorJoin",
				"id":      1,
			},
		})
	testhelpers.ReadMessages(conn, 1, nil)
	assert.True(t, testhelpers.ReadJSON(conn)["success"].(bool))

	//Send ChatMessages
	conn.WriteJSON(
		map[string]interface{}{
			"id": "1",
			"data": map[string]interface{}{
				"request": "getSocketInfo",
			},
		})

	recv := testhelpers.ReadJSON(conn)
	data := recv["data"].(map[string]interface{})
	assert.Equal(t, len(data["rooms"].([]interface{})), 2)
}

func TestChatSend(t *testing.T) {
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
			assert.Equal(t, message["data"].(map[string]interface{})["player"].(map[string]interface{})["steamid"], steamid)
		}
	}

}

// func BenchmarkChatSend(b *testing.B) {
// 	server := testhelpers.StartServer(testhelpers.NewSockets())
// 	defer server.Close()

// 	steamid := strconv.Itoa(rand.Int())
// 	client := testhelpers.NewClient()
// 	testhelpers.Login(steamid, client)
// 	conn, err := testhelpers.ConnectWS(client)
// 	defer conn.Close()
// 	assert.NoError(t, err)
// 	_, err = testhelpers.ReadMessages(conn, testhelpers.InitMessages, nil)
// 	assert.NoError(t, err)

// }

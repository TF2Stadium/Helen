package handler_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
	testhelpers.CleanupDB()
}

func TestLobbyCreate(t *testing.T) {
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
			"password":            "",
			"steamGroupWhitelist": nil,
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
	testhelpers.CleanupDB()
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
				"request":             "lobbyCreate",
				"map":                 "cp_badlands",
				"type":                "6s",
				"league":              "etf2l",
				"server":              strconv.Itoa(i + rand.Int()),
				"rconpwd":             "testerino",
				"whitelistID":         123,
				"mumbleRequired":      true,
				"steamGroupWhitelist": nil,
			}}
		argsChan <- newLobby{args, i}

	}

	wg.Wait()
}

func TestLobbyJoin(t *testing.T) {
	testhelpers.CleanupDB()
	server := testhelpers.StartServer(testhelpers.NewSockets())
	defer server.Close()

	steamid := strconv.Itoa(rand.Int())
	client := testhelpers.NewClient()
	testhelpers.Login(steamid, client)
	player, tperr := models.GetPlayerBySteamId(steamid)
	assert.NoError(t, tperr)
	conn, err := testhelpers.ConnectWS(client)
	assert.NoError(t, err)

	defer conn.Close()

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
	testhelpers.ReadMessages(conn, 2, nil)

	args = map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbyJoin",
			"id":      1,
			"class":   "scout1",
			"team":    "red",
		}}

	conn.WriteJSON(args)

	messages, err := testhelpers.ReadMessages(conn, 3, nil)
	for _, message := range messages {
		_, ok := message["success"]
		if ok {
			assert.True(t, message["success"].(bool))
		} else {
			_, ok := message["request"]
			assert.True(t, ok)
			switch message["request"].(string) {
			case "lobbyJoined", "lobbyListData":
				break
			default:
				t.Logf("Shouldn't be getting request %s", message["request"].(string))
				t.Fail()
			}
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

	args = map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "getSocketInfo",
		},
	}
	conn.WriteJSON(args)
	messages, _ = testhelpers.ReadMessages(conn, 1, nil)
	rooms := (messages[0]["data"].(map[string]interface{}))["rooms"].([]interface{})
	assert.Equal(t, len(rooms), 2)

	for _, room := range rooms {
		switch room.(string) {
		case "0_public", "1_private":
			break
		default:
			t.Logf("Client shouldn't be in room %s", room.(string))
			t.Fail()
		}
	}

}

func TestSpectatorJoin(t *testing.T) {
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
	testhelpers.ReadMessages(conn, 2, nil)

	conn.WriteJSON(
		map[string]interface{}{
			"id": "1",
			"data": map[string]interface{}{
				"request": "lobbySpectatorJoin",
				"id":      1,
			},
		})
	messages, _ := testhelpers.ReadMessages(conn, 2, nil)
	for _, message := range messages {
		success, ok := message["success"]
		if ok {
			assert.True(t, success.(bool))
			continue
		}
		assert.Equal(t, message["request"].(string), "lobbyData")
	}
	var spec int
	db.DB.Table("spectators_players_lobbies").Where("lobby_id = 1").Count(&spec)
	assert.Equal(t, spec, 1)

	//Send ChatMessages
	conn.WriteJSON(
		map[string]interface{}{
			"id": "1",
			"data": map[string]interface{}{
				"request": "getSocketInfo",
			},
		})

	messages, err = testhelpers.ReadMessages(conn, 1, nil)
	assert.NoError(t, err)
	data := messages[0]["data"].(map[string]interface{})
	assert.Equal(t, len(data["rooms"].([]interface{})), 2)

	conn.WriteJSON(
		map[string]interface{}{
			"id": "1",
			"data": map[string]interface{}{
				"request": "lobbySpectatorLeave",
				"id":      1,
			},
		})
	messages, err = testhelpers.ReadMessages(conn, 2, nil)
	assert.NoError(t, err)
}

//Send LobbySpectatorJoin AND LobbyJoin, the way the frontend does it
func TestActualLobbyJoin(t *testing.T) {
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

	var args map[string]interface{}

	args = map[string]interface{}{
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

	testhelpers.ReadMessages(conn, 2, nil)

	args = map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbySpectatorJoin",
			"id":      1,
		},
	}
	conn.WriteJSON(args)
	testhelpers.ReadMessages(conn, 2, nil)

	args = map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbyJoin",
			"id":      1,
			"team":    "blu",
			"class":   "scout1",
		},
	}
	conn.WriteJSON(args)
	testhelpers.ReadMessages(conn, 4, nil)
	var spec int
	db.DB.Table("spectators_players_lobbies").Where("lobby_id = 1").Count(&spec)
	assert.Equal(t, spec, 0)

	args = map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "getSocketInfo",
		},
	}
	conn.WriteJSON(args)
	messages, _ := testhelpers.ReadMessages(conn, 1, nil)
	rooms := (messages[0]["data"].(map[string]interface{}))["rooms"].([]interface{})
	assert.Equal(t, len(rooms), 3)

	for _, room := range rooms {
		switch room.(string) {
		case "0_public", "1_public", "1_private":
			break
		default:
			t.Logf("Client shouldn't be in room %s", room.(string))
			t.Fail()
		}
	}
}

func TestLobbyClose(t *testing.T) {
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

	testhelpers.SocketCreateLobby(conn)
	testhelpers.SocketJoinLobby(conn)

	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbyClose",
			"id":      1,
		},
	}
	conn.WriteJSON(args)
	messages, err := testhelpers.ReadMessages(conn, 5, nil)
	assert.NoError(t, err)

	for _, message := range messages {
		success, ok := message["success"]
		if ok {
			assert.True(t, success.(bool))
			continue
		}
		switch message["request"].(string) {
		case "lobbyLeft", "lobbyClosed", "lobbyData", "lobbyListData":
			continue
		default:
			t.Fatalf("Shouldn't be getting request %s: %v", message["request"].(string), message)
		}
	}

	args = map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbyLeave",
			"id":      1,
		},
	}
	conn.WriteJSON(args)
	messages, _ = testhelpers.ReadMessages(conn, 1, nil)
	assert.Equal(t, len(messages), 1)
	assert.False(t, messages[0]["success"].(bool))

	args = map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbyJoin",
			"id":      1,
			"team":    "red",
			"class":   "scout1",
		},
	}
	conn.WriteJSON(args)
	messages, _ = testhelpers.ReadMessages(conn, 1, nil)
	assert.Equal(t, len(messages), 1)
	assert.False(t, messages[0]["success"].(bool))

}

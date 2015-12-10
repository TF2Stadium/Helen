// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers_test

import (
	"io/ioutil"
	"math/rand"
	"net/url"
	"strconv"
	"testing"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/testhelpers"
	"github.com/TF2Stadium/wsevent"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
	testhelpers.CleanupDB()
}

func TestLogin(t *testing.T) {
	server := testhelpers.StartServer(wsevent.NewServer(), wsevent.NewServer())
	defer server.Close()

	var count int

	steamid := strconv.Itoa(rand.Int())
	client := testhelpers.NewClient()

	resp, err := testhelpers.Login(steamid, client)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	bytes, _ := ioutil.ReadAll(resp.Body)
	t.Log(string(bytes))

	player, tperr := models.GetPlayerBySteamId(steamid)
	assert.NoError(t, tperr)
	assert.NotNil(t, player)
	assert.Equal(t, player.SteamId, steamid)

	assert.Nil(t, db.DB.Table("http_sessions").Count(&count).Error)
	assert.NotEqual(t, count, 0)
	assert.NotNil(t, client.Jar)
}

func TestWS(t *testing.T) {
	server := testhelpers.StartServer(wsevent.NewServer(), wsevent.NewServer())
	defer server.Close()

	steamid := strconv.Itoa(rand.Int())
	client := testhelpers.NewClient()

	resp, err := testhelpers.Login(steamid, client)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	addr, _ := url.Parse("http://localhost:8080/")
	client.Jar.Cookies(addr)
	conn, err := testhelpers.ConnectWS(client)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	for i := 0; i < testhelpers.InitMessages; i++ {
		_, data, err := conn.ReadMessage()
		assert.NoError(t, err)
		t.Log(string(data))
	}
}

func BenchmarkWS(b *testing.B) {
	server := testhelpers.StartServer(wsevent.NewServer(), wsevent.NewServer())
	defer server.Close()

	for i := 0; i < b.N; i++ {
		steamid := strconv.Itoa(rand.Int())
		client := testhelpers.NewClient()
		_, err := testhelpers.Login(steamid, client)
		if err != nil {
			b.Error(err)
		}

		addr, _ := url.Parse("http://localhost:8080/")
		client.Jar.Cookies(addr)
		conn, err := testhelpers.ConnectWS(client)
		if err != nil {
			b.Error(err)
		}

		for i := 0; i < 5; i++ {
			_, _, err := conn.ReadMessage()
			if err != nil {
				b.Error(err)
			}
		}

	}
}

func TestSocketInfo(t *testing.T) {
	server := testhelpers.StartServer(testhelpers.NewSockets())
	defer server.Close()

	client := testhelpers.NewClient()
	testhelpers.Login(strconv.Itoa(rand.Int()), client)
	conn, err := testhelpers.ConnectWS(client)
	assert.NoError(t, err)

	_, err = testhelpers.ReadMessages(conn, testhelpers.InitMessages, t)
	assert.NoError(t, err)

	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "getSocketInfo",
		},
	}

	reply, err := testhelpers.EmitJSONWithReply(conn, args)
	assert.NoError(t, err)
	//assert.Equal(t, reply["rooms"].([]interface{})[0].(string), "0_public")
	t.Logf("%v", reply)

	conn.Close()
}

func TestLobbyCreate(t *testing.T) {
	server := testhelpers.StartServer(testhelpers.NewSockets())
	defer server.Close()

	steamid := strconv.Itoa(rand.Int())
	client := testhelpers.NewClient()
	testhelpers.Login(steamid, client)
	conn, err := testhelpers.ConnectWS(client)
	assert.NoError(t, err)
	_, err = testhelpers.ReadMessages(conn, testhelpers.InitMessages, t)
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

	reply, err := testhelpers.EmitJSONWithReply(conn, args)
	assert.NoError(t, err)
	assert.True(t, reply["success"].(bool))
	id := uint(reply["data"].(map[string]interface{})["id"].(float64))
	t.Logf("%v", reply)

	lobby, err := models.GetLobbyById(id)
	assert.NoError(t, err)
	assert.Equal(t, lobby.CreatedBySteamID, steamid)
	conn.Close()
}

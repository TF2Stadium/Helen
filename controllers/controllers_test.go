// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers_test

import (
	"io/ioutil"
	"math/rand"
	"net/url"
	"strconv"
	//"sync"
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

	testhelpers.ReadMessages(conn, testhelpers.InitMessages, t)
}

func BenchmarkWS(b *testing.B) {
	server := testhelpers.StartServer(wsevent.NewServer(), wsevent.NewServer())
	defer server.Close()
	//wg := new(sync.WaitGroup)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//go func() {
		//wg.Add(1)
		//defer wg.Done()

		steamid := strconv.Itoa(rand.Int())
		client := testhelpers.NewClient()
		_, err := testhelpers.Login(steamid, client)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}

		conn, err := testhelpers.ConnectWS(client)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}

		for i := 0; i < 5; i++ {
			_, _, err := conn.ReadMessage()
			if err != nil {
				b.Error(err)
				b.FailNow()
			}
		}
		//}()
	}
	//wg.Wait()
}

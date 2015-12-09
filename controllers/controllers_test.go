// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"testing"

	"github.com/TF2Stadium/Helen/config"
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

	auth := wsevent.NewServer()
	noauth := wsevent.NewServer()
	config.Constants.MockupAuth = true
	SetupHTTPRoutes(auth, noauth)
	go func() {
		helpers.Logger.Fatal(http.ListenAndServe(":8080", nil))
	}()
}

func TestLogin(t *testing.T) {
	var count int

	steamid := strconv.Itoa(rand.Int())
	client := new(http.Client)
	client.Jar, _ = cookiejar.New(nil)

	resp, err := testhelpers.Login(steamid, client)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	player, tperr := models.GetPlayerBySteamId(steamid)
	assert.NoError(t, tperr)
	assert.Equal(t, player.SteamId, steamid)

	assert.Nil(t, db.DB.Table("http_sessions").Count(&count).Error)
	assert.NotEqual(t, count, 0)
	assert.NotNil(t, client.Jar)
}

func TestWS(t *testing.T) {
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

	for i := 0; i < 5; i++ {
		_, data, err := conn.ReadMessage()
		assert.NoError(t, err)
		t.Log(string(data))
	}
}

func BenchmarkWS(b *testing.B) {
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

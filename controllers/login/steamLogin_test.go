// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package login_test

import (
	"io/ioutil"
	"math/rand"
	"net/url"
	"strconv"
	//"sync"
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
	testhelpers.StartServer()
}

func TestLogin(t *testing.T) {
	var count int

	steamid := strconv.Itoa(rand.Int())
	client := testhelpers.NewClient()

	resp, err := testhelpers.Login(steamid, client)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	bytes, _ := ioutil.ReadAll(resp.Body)
	t.Log(string(bytes))

	player, tperr := models.GetPlayerBySteamID(steamid)
	assert.NoError(t, tperr)
	assert.NotNil(t, player)
	assert.Equal(t, player.SteamID, steamid)

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

	testhelpers.ReadMessages(conn, testhelpers.InitMessages, t)
}

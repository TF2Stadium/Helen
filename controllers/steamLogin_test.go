// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	//"github.com/TF2Stadium/Helen/routes"
	"github.com/TF2Stadium/Helen/testhelpers"
	"github.com/stretchr/testify/assert"
)

var client = new(http.Client)

func init() {
	testhelpers.CleanupDB()
	database.Init()
	migrations.Do()
	stores.SetupStores()
	helpers.InitLogger()

	//routes.SetupHTTPRoutes(,)
	go func() {
		helpers.Logger.Fatal(http.ListenAndServe(":8080", nil))
	}()
}

func TestLogin(t *testing.T) {
	steamid := strconv.Itoa(rand.Int())

	addr, _ := url.Parse("http://localhost:8080/startMockLogin/" + steamid)
	resp, err := client.Do(&http.Request{Method: "GET", URL: addr})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	player, tperr := models.GetPlayerBySteamId(steamid)
	assert.Nil(t, tperr)
	assert.Equal(t, player.SteamId, steamid)
}

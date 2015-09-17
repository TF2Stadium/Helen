// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"github.com/TF2Stadium/Helen/config"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func init() {
	config.Constants.SteamApiMockUp = true
}

func TestFakeSocket(t *testing.T) {
	log.Println("test started")
	socket := NewFakeSocket()
	assert.NotEqual(t, "", socket.Id())

	rooms := socket.Rooms()
	assert.Equal(t, 0, len(rooms))

	socket.Join("0")
	socket.Join("123490")

	rooms2 := socket.Rooms()

	assert.Contains(t, rooms2, "0")
	assert.Contains(t, rooms2, "123490")

	socket.Leave("0")
	socket.Leave("123490")
	rooms3 := socket.Rooms()
	assert.Equal(t, 0, len(rooms3))
}

func TestFakeSocketBroadcast(t *testing.T) {
	socket := NewFakeSocket()
	socket.Join("0")
	rooms4 := socket.Rooms()
	assert.Equal(t, []string{"0"}, rooms4)

	testCounter := 0

	//	socket2 := NewFakeSocket()

	f := func(in string) string {
		testCounter += 1
		return "{}"
	}

	socket.On("test", f)
	//	socket2.On("test", f)

	socket.SimRequest("test", "o hai")
	assert.Equal(t, 1, testCounter)

	socket.Emit("test2", "{\"data\": 2}")
	event, jsonD := socket.GetNextMessage()
	assert.NotNil(t, jsonD)
	assert.Equal(t, "test2", event)
	assert.Equal(t, 2, jsonD.Get("data").MustInt())

	event, jsonD = socket.GetNextMessage()
	assert.Nil(t, jsonD)

	socket2 := NewFakeSocket()
	socket.Join("0")
	socket2.Join("0")

	socket2.BroadcastTo("0", "test2", "{\"data\": 2}")
	event, jsonD = socket.GetNextMessage()
	assert.NotNil(t, jsonD)
	assert.Equal(t, "test2", event)
	assert.Equal(t, 2, jsonD.Get("data").MustInt())

	event, jsonD = socket2.GetNextMessage()
	assert.Nil(t, jsonD)

	FakeSocketServer.BroadcastTo("0", "test2", "{\"data\": 2}")

	event, jsonD = socket.GetNextMessage()
	assert.NotNil(t, jsonD)
	assert.Equal(t, "test2", event)
	assert.Equal(t, 2, jsonD.Get("data").MustInt())

	event, jsonD = socket2.GetNextMessage()
	assert.NotNil(t, jsonD)
	assert.Equal(t, "test2", event)
	assert.Equal(t, 2, jsonD.Get("data").MustInt())
}

func TestFakeSocketAuthentication(t *testing.T) {
	socket := NewFakeSocket()

	player, _ := models.NewPlayer("1234")
	//	player.Save()

	socket.On("test", func(in string) {
		assert.False(t, chelpers.IsLoggedInSocket(socket.Id()))
	})
	socket.SimRequest("test", "asdf")

	socket.FakeAuthenticate(player)

	socket.On("test", func(in string) {
		assert.True(t, chelpers.IsLoggedInSocket(socket.Id()))
	})
	socket.SimRequest("test", "asdf")
}

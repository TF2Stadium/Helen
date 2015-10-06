// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"errors"
	"fmt"
	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"
)

// taken from http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type fakeSocketServer struct {
	sync.Mutex
	sockets []*fakeSocket
	rooms   map[string]map[string]socketio.Socket
}

func (f *fakeSocketServer) BroadcastTo(room string, message string, args ...interface{}) {
	f.Lock()
	defer f.Unlock()
	for _, rooms := range f.rooms {
		socket, ok := rooms[room]
		if !ok {
			continue
		}
		socket.Emit(message, args[0].(string))
	}
	return
}

var FakeSocketServer = fakeSocketServer{sockets: nil, rooms: make(map[string]map[string]socketio.Socket)}

type message struct {
	event   string
	content string
}

type fakeSocket struct {
	id            string
	receivedQueue chan message
	eventHandlers map[string]interface{}
	server        *fakeSocketServer
}

func NewFakeSocket() *fakeSocket {
	FakeSocketServer.Lock()
	defer FakeSocketServer.Unlock()
	so := &fakeSocket{
		id:            RandSeq(5),
		receivedQueue: make(chan message, 100),
		eventHandlers: make(map[string]interface{}),
		server:        &FakeSocketServer,
	}

	FakeSocketServer.sockets = append(FakeSocketServer.sockets, so)
	return so
}

func (f *fakeSocket) GetNextMessage() (string, *simplejson.Json) {
	// broadcasts are sent in a different thread so wait a bit to make sure they are received
	time.Sleep(500 * time.Microsecond)
	select {
	case s := <-f.receivedQueue:
		json, _ := simplejson.NewJson([]byte(s.content))
		return s.event, json
	default:
		return "", nil
	}
}

func (f *fakeSocket) GetNextNamedMessage(name string) *simplejson.Json {
	for {
		msg, res := f.GetNextMessage()
		if res == nil {
			return nil
		} else if msg == name {
			return res
		}
	}
}

func (f *fakeSocket) Id() string {
	return f.id
}

func (f *fakeSocket) SimRequest(event string, args string) (*simplejson.Json, error) {
	fn, ok := f.eventHandlers[event]
	if !ok {
		return nil, errors.New("event not recognized")
	}

	fnAsserted, ok2 := fn.(func(string) string)
	if ok2 {
		return simplejson.NewJson([]byte(fnAsserted(args)))
	}

	// doesn't have to return string
	fnAsserted2 := fn.(func(string))
	fnAsserted2(args)
	return nil, nil
}

func (f *fakeSocket) Join(room string) error {
	f.server.Lock()
	defer f.server.Unlock()
	arr, ok := f.server.rooms[f.Id()]
	if !ok {
		arr = make(map[string]socketio.Socket)
		f.server.rooms[f.Id()] = arr
	}
	arr[room] = f
	return nil
}

func (f *fakeSocket) Leave(room string) error {
	f.server.Lock()
	defer f.server.Unlock()
	arr, ok := f.server.rooms[f.Id()]
	if ok {
		delete(arr, room)
	}
	return nil
}

func (f *fakeSocket) Rooms() []string {
	f.server.Lock()
	defer f.server.Unlock()
	var rooms []string
	arr, ok := f.server.rooms[f.Id()]
	if !ok {
		return rooms
	}
	for room, _ := range arr {
		rooms = append(rooms, room)
	}
	return rooms
}

func (f *fakeSocket) On(message string, fn interface{}) error {
	f.eventHandlers[message] = fn
	return nil
}

func (f *fakeSocket) Emit(event string, args ...interface{}) error {
	// ASSUMES A STRING IS EMITTED
	f.receivedQueue <- message{event, args[0].(string)}
	return nil
}

func (f *fakeSocket) BroadcastTo(room, message string, args ...interface{}) error {
	for sid, rooms := range f.server.rooms {
		if sid == f.Id() {
			continue
		}

		socket, ok := rooms[room]
		if !ok {
			continue
		}
		socket.Emit(message, args[0].(string))
	}
	return nil
}

type fakeReader struct{}

func (f *fakeReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (f *fakeSocket) Request() *http.Request {
	r, _ := http.NewRequest("GET", "", &fakeReader{})
	return r
}

func (f *fakeSocket) FakeAuthenticate(player *models.Player) *http.Request {
	session := &sessions.Session{
		ID:      RandSeq(5),
		Values:  make(map[interface{}]interface{}),
		Options: nil,
		IsNew:   false,
	}

	session.Values["id"] = fmt.Sprint(player.ID)
	session.Values["steam_id"] = fmt.Sprint(player.SteamId)
	session.Values["role"] = player.Role

	broadcaster.SetSocket(player.SteamId, f)
	stores.SetSocketSession(f.Id(), session)
	return nil
}

func SetupFakeSockets() {
	broadcaster.Init(&FakeSocketServer)
}

func UnpackSuccessResponse(t *testing.T, r *simplejson.Json) *simplejson.Json {
	assert.NotNil(t, r)
	assert.True(t, r.Get("success").MustBool(), fmt.Sprintf(
		"A request should have been successful but it failed: %s", r.Get("message").MustString()))
	return r.Get("data")
}

func UnpackFailureResponse(t *testing.T, r *simplejson.Json) (int, string) {
	assert.NotNil(t, r)
	assert.False(t, r.Get("success").MustBool(), "A request should have failed but was successful")
	return r.Get("code").MustInt(), r.Get("message").MustString()
}

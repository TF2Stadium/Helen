// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/TF2Stadium/wsevent"
	"github.com/gorilla/websocket"
)

func NewClient() (client *http.Client) {
	client = new(http.Client)
	client.Jar, _ = cookiejar.New(nil)
	return
}

func Login(steamid string, client *http.Client) (*http.Response, error) {
	addr, _ := url.Parse("http://localhost:8080/startMockLogin/" + steamid)
	return client.Do(&http.Request{Method: "GET", URL: addr})
}

func ConnectWS(client *http.Client) (*websocket.Conn, error) {
	ws := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/websocket/"}
	domain := &url.URL{Scheme: "http", Host: "localhost:8080"}

	resp, err := client.Head(domain.String())
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(ws.String(), resp.Header)
	return conn, err
}

func EmitJSONWithReply(client *wsevent.Client, v interface{}) {
	bytes, _ := json.Marshal(v)
	client.Emit(string(bytes))
}

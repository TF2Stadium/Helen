// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/TF2Stadium/Helen/controllers"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	"github.com/TF2Stadium/Helen/controllers/socket"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/gorilla/websocket"
)

const InitMessages int = 5

type SuffixList struct{}

var (
	options = &cookiejar.Options{PublicSuffixList: SuffixList{}}
)

func (SuffixList) PublicSuffix(_ string) string {
	return ""
}

func (SuffixList) String() string {
	return ""
}

var DefaultTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 5 * time.Second,
	}).Dial,
}

func NewClient() (client *http.Client) {
	client = new(http.Client)
	client.Transport = DefaultTransport
	DefaultTransport.CloseIdleConnections()
	client.Jar, _ = cookiejar.New(options)
	return
}

func Login(steamid string, client *http.Client) (*http.Response, error) {
	addr, _ := url.Parse("http://localhost:8080/startMockLogin/" + steamid)
	return client.Do(&http.Request{Method: "GET", URL: addr})
}

func ConnectWS(client *http.Client) (*websocket.Conn, error) {
	ws := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/websocket/"}
	domain := &url.URL{Scheme: "http", Host: "localhost:8080"}

	if len(client.Jar.Cookies(domain)) == 0 {
		return nil, errors.New("Client cookiejar has no cookies D:")
	}

	header := http.Header{"Cookie": []string{client.Jar.Cookies(domain)[0].String()}}

	conn, _, err := websocket.DefaultDialer.Dial(ws.String(), header)
	return conn, err
}

func LoginAndConnectWS() (string, *websocket.Conn, *http.Client, error) {
	steamid := strconv.Itoa(rand.Int())
	client := NewClient()

	_, err := Login(steamid, client)
	if err != nil {
		return "", nil, nil, err
	}

	conn, err := ConnectWS(client)
	if err != nil {
		return "", nil, nil, err
	}

	_, err = ReadMessages(conn, InitMessages, nil)

	return steamid, conn, client, err
}

func EmitJSONWithReply(conn *websocket.Conn, req map[string]interface{}) (map[string]interface{}, error) {
	if err := conn.WriteJSON(req); err != nil {
		return nil, errors.New("Error while marshing request: " + err.Error())
	}

	resp := make(map[string]interface{})

	if err := conn.ReadJSON(&resp); err != nil {
		return nil, errors.New("Error while marshing response: " + err.Error())
	}

	return resp["data"].(map[string]interface{}), nil
}

func resetServers() {
	socket.NewServers()
	broadcaster.Init(socket.AuthServer, socket.UnauthServer)
}

func StartServer() *httptest.Server {
	resetServers()
	var mux = http.NewServeMux()
	mux.HandleFunc("/", controllers.MainHandler)
	mux.HandleFunc("/openidcallback", controllers.LoginCallbackHandler)
	mux.HandleFunc("/startLogin", controllers.LoginHandler)
	mux.HandleFunc("/startMockLogin/", controllers.MockLoginHandler)
	mux.HandleFunc("/logout", controllers.LogoutHandler)
	mux.HandleFunc("/websocket/", controllers.SocketHandler)

	l, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		for err != nil {
			l, err = net.Listen("tcp", "localhost:8080")
		}
	}

	server := &httptest.Server{Listener: l, Config: &http.Server{Handler: mux}}
	go server.Start()
	return server
}

func ReadMessages(conn *websocket.Conn, n int, t *testing.T) ([]map[string]interface{}, error) {
	var messages []map[string]interface{}
	for i := 0; i < n; i++ {
		data := ReadJSON(conn)
		messages = append(messages, data)

		if t != nil {
			bytes, _ := json.MarshalIndent(data, "", "  ")
			t.Logf("%s", string(bytes))
		}
	}

	return messages, nil
}

func ReadJSON(conn *websocket.Conn) map[string]interface{} {
	reply := make(map[string]interface{})

	err := conn.ReadJSON(&reply)
	if err != nil {
		helpers.Logger.Error(err.Error())
	}

	return reply["data"].(map[string]interface{})
}

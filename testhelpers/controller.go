// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"encoding/json"
	"errors"
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
	"github.com/TF2Stadium/wsevent"
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

func EmitJSONWithReply(conn *websocket.Conn, req map[string]interface{}) (map[string]interface{}, error) {
	if err := conn.WriteJSON(req); err != nil {
		return nil, errors.New("Error while marshing request: " + err.Error())
	}

	var resp struct {
		Id   string
		Data string
	}

	if err := conn.ReadJSON(&resp); err != nil {
		return nil, errors.New("Error while marshing response: " + err.Error())
	}

	data := make(map[string]interface{})
	str, _ := strconv.Unquote(resp.Data)

	if err := json.Unmarshal([]byte(str), &data); err != nil {
		return nil, errors.New("Error while marshing response data: " + err.Error())
	}

	return data, nil
}

func StartServer(auth *wsevent.Server, noauth *wsevent.Server) *httptest.Server {
	var mux = http.NewServeMux()
	mux.HandleFunc("/", controllers.MainHandler)
	mux.HandleFunc("/openidcallback", controllers.LoginCallbackHandler)
	mux.HandleFunc("/startLogin", controllers.LoginHandler)
	mux.HandleFunc("/startMockLogin/", controllers.MockLoginHandler)
	mux.HandleFunc("/logout", controllers.LogoutHandler)
	mux.HandleFunc("/chatlogs/", controllers.GetChatLogs)
	mux.HandleFunc("/websocket/", controllers.Sockets{auth, noauth}.SocketHandler)

	broadcaster.Init(auth, noauth)

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

	switch reply["data"].(type) {
	case string:
		//reply to a request
		data := make(map[string]interface{})
		str, _ := strconv.Unquote(reply["data"].(string))

		json.Unmarshal([]byte(str), &data)
		return data

	case map[string]interface{}:
		//request from server
		dataStr := reply["data"].(map[string]interface{})["data"].(string)
		dataMap := make(map[string]interface{})
		json.Unmarshal([]byte(dataStr), &dataMap)

		reply["data"].(map[string]interface{})["data"] = dataMap
		return reply["data"].(map[string]interface{})

	default:
		panic("Received invalid JSON")
	}

}

func getEvent(data []byte) string {
	var js struct {
		Request string
	}
	json.Unmarshal(data, &js)
	return js.Request
}

func NewSockets() (*wsevent.Server, *wsevent.Server) {
	auth := wsevent.NewServer()
	noauth := wsevent.NewServer()

	auth.Extractor = getEvent
	noauth.Extractor = getEvent

	socket.ServerInit(auth, noauth)
	return auth, noauth
}

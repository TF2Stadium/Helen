package controllers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/gorilla/websocket"
)

var (
	wsConnTime = new(int64)
)

type stats struct {
	WSConnTime int64
}

func getStats() stats {
	return stats{atomic.LoadInt64(wsConnTime)}
}

func Health(w http.ResponseWriter, r *http.Request) {
	stats := getStats()
	json.NewEncoder(w).Encode(stats)
}

func StartHealthCheck() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			dur := int64(webSocketTime())
			atomic.StoreInt64(wsConnTime, dur)
		}
	}()
}

//measure the amount of time it takes for a
//websocket connection to initialize
func webSocketTime() time.Duration {
	start := time.Now()
	u := url.URL{
		Scheme: "ws",
		Host:   config.Constants.ListenAddress,
		Path:   "websocket/",
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Error("Cannot make connection to WebSocket endpoint: ", err)
		return 0
	}

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	for i := 0; i < 4; i++ {
		js := make(map[string]interface{})
		err = conn.ReadJSON(&js)
		if err != nil {
			logrus.Error("Error while connecting to websocket endpoint: ", err)
			return 0
		}
	}
	end := time.Now()
	conn.Close()

	return end.Sub(start)
}

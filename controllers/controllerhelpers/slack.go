package controllerhelpers

import (
	"fmt"
	"net/http"
	"bytes"
	"sync"
	"time"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
)

type message struct {
	Name    string
	SteamID string
	Message string
}

type SlackMessage struct {
	Text string `json:"text"`
}

var messages = make(chan message, 10)
var once = new(sync.Once)

func slackBroadcaster() {
	for {
		m := <-messages
		final := fmt.Sprintf("<https://steamcommunity.com/profiles/%s|%s>: %s", m.SteamID, m.Name, m.Message)
		payload, _ := json.Marshal(SlackMessage{final})
		_, err := http.Post(config.Constants.SlackbotURL, "application/json",
			bytes.NewReader(payload))

		if err != nil {
			logrus.Error(err.Error())
		}

		time.Sleep(time.Second * 1)
	}
}

func SendToSlack(msg, name, steamid string) {
	if config.Constants.SlackbotURL == "" {
		return
	}
	go once.Do(slackBroadcaster)

	messages <- message{name, steamid, msg}

}

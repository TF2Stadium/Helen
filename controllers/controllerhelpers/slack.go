package controllerhelpers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
)

type message struct {
	Name    string
	SteamID string
	Message string
}

var messages = make(chan message, 10)
var once = new(sync.Once)

func slackBroadcaster() {
	for {
		m := <-messages
		final := fmt.Sprintf("<https://steamcommunity.com/profiles/%s|%s>: %s", m.SteamID, m.Name, m.Message)
		_, err := http.Post(config.Constants.SlackbotURL, "text/plain",
			strings.NewReader(final))

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

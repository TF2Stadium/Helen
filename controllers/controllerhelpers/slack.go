package controllerhelpers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
)

type message struct {
	Name    string
	SteamID string
	Message string
}

var messages = make(chan message, 10)

func SlackBroadcaster() {
	if config.Constants.SlackbotURL == "" {
		return
	}

	for {
		m := <-messages
		final := fmt.Sprintf("<https://steamcommunity.com/profiles/%s|%s>: %s", m.SteamID, m.Name, m.Message)
		_, err := http.Post(config.Constants.SlackbotURL, "text/plain",
			strings.NewReader(final))

		if err != nil {
			helpers.Logger.Error(err.Error())
		}

		time.Sleep(time.Second * 1)
	}
}

func SendToSlack(msg, name, steamid string) {
	if config.Constants.SlackbotURL == "" {
		return
	}

	messages <- message{name, steamid, msg}
}

package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/chat"
)

func TFTVStreamStatusUpdater() {
	ticker := time.NewTicker(5 * time.Minute)
	u := &url.URL{
		Scheme: "https",
		Host:   "api.twitch.tv",
		Path:   "kraken/streams",
	}
	values := u.Query()
	values.Set("game", "Team Fortress 2")
	values.Set("channel", "teamfortresstv")
	values.Set("stream_type", "live")
	u.RawQuery = values.Encode()

	var reply struct {
		Total   int `json:"_total"`
		Streams []struct {
			Channel struct {
				Status string `json:"status"`
			} `json:"channel"`
		} `json:"streams"`
	}
	var streaming bool

	for {
		req, _ := http.NewRequest("GET", u.String(), nil)
		req.Header.Set("Accept", "application/vnd.twitchtv.v3+json")
		resp, err := helpers.HTTPClient.Do(req)
		if err != nil {
			logrus.Error(err)
			continue
		}

		json.NewDecoder(resp.Body).Decode(&reply)
		if reply.Total != 0 && !streaming {
			str := fmt.Sprintf(`twitch.tv/teamfortresstv is live with "%s"`, reply.Streams[0].Channel.Status)
			message := chat.NewBotMessage(str, 0)
			message.Send()
			streaming = true
		} else {
			streaming = false
		}
		<-ticker.C
	}
}

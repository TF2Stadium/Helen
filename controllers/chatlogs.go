package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
)

func GetChatLogs(w http.ResponseWriter, r *http.Request) {
	var messages []*models.ChatMessage
	logs := "<body>\n"

	steamid := strings.Index(r.URL.Path, "steamid/")
	room := strings.Index(r.URL.Path, "room/")

	if steamid != -1 {
		var err error

		steamid := r.URL.Path[strings.Index(r.URL.Path, "steamid/")+8:]
		player, tperr := models.GetPlayerBySteamId(steamid)
		if tperr != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		messages, err = models.GetPlayerMessages(player)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

	} else if room != -1 {
		roomstr := r.URL.Path[strings.Index(r.URL.Path, "room/")+5:]
		room, err := strconv.Atoi(roomstr)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		messages, err = models.GetRoomMessages(room)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}

	format := "<font color=\"red\">[%s]</font> <a href=\"https://steamcommunity.com/profiles/%s\">%s</a>: %s<br>\n"
	if steamid != -1 {
		format = "<font color=\"red\">[%s]</font> <a href=\"https://steamcommunity.com/profiles/%s\">%s</a>: %s<br>\n"
	}

	var player models.Player
	prevRoom := -1
	//format := "%s: %s\t[%s]\n"
	for _, message := range messages {
		if prevRoom != message.Room {
			logs += fmt.Sprintf("<font color=\"blue\"> Room #%d </font><br>\n", message.Room)
		}

		prevRoom = message.Room
		db.DB.First(&player, message.PlayerID)

		if steamid != -1 {
			logs += fmt.Sprintf(format, message.CreatedAt.Format(time.RFC822), player.SteamId, player.Name, message.Message)
			continue
		}

		logs += fmt.Sprintf(format, message.CreatedAt.Format(time.RFC822), player.SteamId, player.Name, message.Message)
	}

	logs += "</body>"
	fmt.Fprint(w, logs)
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
	"github.com/bitly/go-simplejson"
)

type chatMessage struct {
	Timestamp int64                `json:"timestamp"`
	Message   string               `json:"message"`
	Room      int                  `json:"room"`
	Player    models.PlayerSummary `json:"player"`
}

func ChatSend(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}
	var args struct {
		Message *string `json:"message"`
		Room    *int    `json:"room"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
	if tperr != nil {
		bytes, _ := tperr.ErrorJSON().Encode()
		return string(bytes)
	}

	//helpers.Logger.Debug("received chat message: %s %s", *args.Message, player.Name)

	spec := player.IsSpectatingId(uint(*args.Room))
	//Check if player has either joined, or is spectating lobby
	lobbyId, tperr := player.GetLobbyId()
	if *args.Room > 0 {
		if tperr != nil && !spec && lobbyId != uint(*args.Room) {
			bytes, _ := chelpers.BuildFailureJSON("Player is not in the lobby.", 5).Encode()
			return string(bytes)
		}
	} else {
		// else room is the lobby list room
		*args.Room, _ = strconv.Atoi(config.Constants.GlobalChatRoom)
	}

	message := chatMessage{
		Timestamp: time.Now().Unix(),
		Message:   *args.Message,
		Room:      *args.Room,
		Player:    models.DecoratePlayerSummary(player)}

	bytes, _ := json.Marshal(message)
	broadcaster.SendMessageToRoom(fmt.Sprintf("%s_public",
		chelpers.GetLobbyRoom(uint(*args.Room))),
		"chatReceive", string(bytes))

	resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()

	helpers.Logger.Debug("%t", strings.HasPrefix(*args.Message, "!admin"))
	if strings.HasPrefix(*args.Message, "!admin") {
		chelpers.SendToSlack(*args.Message, player.Name, player.SteamId)
	}

	chelpers.LogChat(uint(*args.Room), player.Name, player.SteamId, *args.Message)

	chelpers.AddScrollbackMessage(uint(*args.Room), string(bytes))
	return string(resp)

}

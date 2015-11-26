// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

type chatMessage struct {
	Timestamp int64                `json:"timestamp"`
	Message   string               `json:"message"`
	Room      int                  `json:"room"`
	Player    models.PlayerSummary `json:"player"`
	Id        uint32                 `json:"id"`
}

var nextId uint32 = 0

func ChatSend(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}
	var args struct {
		Message *string `json:"message"`
		Room    *int    `json:"room"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
	if tperr != nil {
		return tperr.Encode()
	}

	//helpers.Logger.Debug("received chat message: %s %s", *args.Message, player.Name)

	if *args.Room > 0 {
		spec := player.IsSpectatingId(uint(*args.Room))
		//Check if player has either joined, or is spectating lobby
		lobbyId, tperr := player.GetLobbyId()

		if tperr != nil && !spec && lobbyId != uint(*args.Room) {
			return helpers.NewTPError("Player is not in the lobby.", 5).Encode()
		}
	} else {
		// else room is the lobby list room
		*args.Room, _ = strconv.Atoi(config.Constants.GlobalChatRoom)
	}

	message := chatMessage{
		Timestamp: time.Now().Unix(),
		Message:   *args.Message,
		Room:      *args.Room,
		Player:    models.DecoratePlayerSummary(player),
		Id:        atomic.AddUint32(&nextId, 1),
	}

	bytes, _ := json.Marshal(message)
	broadcaster.SendMessageToRoom(fmt.Sprintf("%s_public",
		chelpers.GetLobbyRoom(uint(*args.Room))),
		"chatReceive", string(bytes))

	if strings.HasPrefix(*args.Message, "!admin") {
		chelpers.SendToSlack(*args.Message, player.Name, player.SteamId)
	}

	chelpers.LogChat(uint(*args.Room), player.Name, player.SteamId, *args.Message)

	chelpers.AddScrollbackMessage(uint(*args.Room), string(bytes))
	return chelpers.EmptySuccessJS
}

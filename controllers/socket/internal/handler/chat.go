// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

type Chat struct{}

func (Chat) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

var lastChatTime = make(map[string]int64)

func (Chat) ChatSend(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr
	}

	steamid := chelpers.GetSteamId(so.Id())
	now := time.Now().Unix()
	if now-lastChatTime[steamid] == 0 {
		return helpers.NewTPError("You're sending messages too quickly", -1)
	}

	player, tperr := models.GetPlayerBySteamId(steamid)

	var args struct {
		Message *string `json:"message"`
		Room    *int    `json:"room"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	lastChatTime[steamid] = now
	if tperr != nil {
		return tperr
	}

	//helpers.Logger.Debug("received chat message: %s %s", *args.Message, player.Name)

	if *args.Room > 0 {
		var count int
		spec := player.IsSpectatingId(uint(*args.Room))
		//Check if player has either joined, or is spectating lobby
		db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", *args.Room, player.ID).Count(&count)

		if !spec && count == 0 {
			return helpers.NewTPError("Player is not in the lobby.", 5)
		}
	} else {
		// else room is the lobby list room
		*args.Room, _ = strconv.Atoi(config.Constants.GlobalChatRoom)
	}
	if len(*args.Message) != 0 && (*args.Message)[0] == '\n' {
		return helpers.NewTPError("Cannot send messages prefixed with newline", 4)
	}

	if len(*args.Message) > 120 {
		return helpers.NewTPError("Message too long", 4)
	}

	message := models.NewChatMessage(*args.Message, *args.Room, player)
	if tperr != nil {
		return tperr
	}
	db.DB.Save(message)
	broadcaster.SendMessageToRoom(fmt.Sprintf("%s_public",
		chelpers.GetLobbyRoom(uint(*args.Room))),
		"chatReceive", message)

	if strings.HasPrefix(*args.Message, "!admin") {
		chelpers.SendToSlack(*args.Message, player.Name, player.SteamId)
	}

	return chelpers.EmptySuccessJS
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/TF2Stadium/Helen/config"
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

func (Chat) ChatSend(so *wsevent.Client, data []byte) interface{} {
	steamid := chelpers.GetSteamId(so.ID)
	now := time.Now().Unix()
	if now-lastChatTime[steamid] == 0 {
		return helpers.NewTPError("You're sending messages too quickly", -1)
	}

	player, tperr := models.GetPlayerBySteamID(steamid)

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
		spec := player.IsSpectatingID(uint(*args.Room))
		//Check if player has either joined, or is spectating lobby
		db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", *args.Room, player.ID).Count(&count)

		if !spec && count == 0 {
			return helpers.NewTPError("Player is not in the lobby.", 5)
		}
	} else {
		// else room is the lobby list room
		*args.Room, _ = strconv.Atoi(config.Constants.GlobalChatRoom)
	}
	switch {
	case len(*args.Message) == 0:
		return helpers.NewTPError("Cannot send an empty message", 4)

	case (*args.Message)[0] == '\n':
		return helpers.NewTPError("Cannot send messages prefixed with newline", 4)

	case len(*args.Message) > 150:
		return helpers.NewTPError("Message too long", 4)
	}

	message := models.NewChatMessage(*args.Message, *args.Room, player)
	if tperr != nil {
		return tperr
	}

	message.Save()

	if strings.HasPrefix(*args.Message, "!admin") {
		chelpers.SendToSlack(*args.Message, player.Name, player.SteamID)
		return chelpers.EmptySuccessJS
	}

	message.Send()

	return chelpers.EmptySuccessJS
}

func (Chat) ChatDelete(so *wsevent.Client, data []byte) interface{} {
	if err := chelpers.CheckPrivilege(so, helpers.ActionDeleteChat); err != nil {
		return err
	}

	var args struct {
		ID   *int  `json:"id"`
		Room *uint `json:"room"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return err
	}

	message := &models.ChatMessage{}
	err := db.DB.Table("chat_messages").Where("room = ? AND id = ?", args.Room, args.ID).First(message).Error
	if message.Bot {
		return helpers.NewTPError("Cannot delete notification messages", -1)
	}
	if err != nil {
		return helpers.NewTPError("Can't find message", -1)
	}

	player, _ := models.GetPlayerByID(message.PlayerID)
	message.Deleted = true
	message.Timestamp = message.CreatedAt.Unix()
	message.Save()
	message.Message = "<deleted>"
	message.Player = models.DecoratePlayerSummary(player)
	message.Player.Tags = append(message.Player.Tags, "deleted")
	message.Send()

	return chelpers.EmptySuccessJS
}

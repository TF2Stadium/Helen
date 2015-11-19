// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"encoding/json"
	"fmt"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

func DebugLobbyReady(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}

	var args struct {
		Id *uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	lobby, _ := models.GetLobbyById(*args.Id)

	var slots []models.LobbySlot
	db.DB.Where("lobby_id = ?", lobby.ID).Find(&slots)
	for _, slot := range slots {
		slot.Ready = true
		db.DB.Save(slot)
	}
	lobby.OnChange(true)

	return chelpers.EmptySuccessJS
}

func DebugRequestLobbyStart(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}

	var args struct {
		Id *uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	lobby, _ := models.GetLobbyByIdServer(*args.Id)
	bytes, _ := json.Marshal(models.DecorateLobbyConnect(lobby))
	room := fmt.Sprintf("%s_private", chelpers.GetLobbyRoom(lobby.ID))
	broadcaster.SendMessageToRoom(room, "lobbyStart", string(bytes))

	return chelpers.EmptySuccessJS
}

func DebugUpdateStatsFilter(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}

	var args struct {
		Id *uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	lobby, tperr := models.GetLobbyById(*args.Id)
	if tperr != nil {
		return tperr.Encode()
	}
	lobby.UpdateStats()

	return chelpers.EmptySuccessJS
}

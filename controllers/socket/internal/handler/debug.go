// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"fmt"

	"encoding/json"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
	"github.com/bitly/go-simplejson"
)

func DebugLobbyReady(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}

	var args struct {
		Id *uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	lobby, _ := models.GetLobbyById(*args.Id)

	var slots []models.LobbySlot
	db.DB.Where("lobby_id = ?", lobby.ID).Find(&slots)
	for _, slot := range slots {
		slot.Ready = true
		db.DB.Save(slot)
	}
	lobby.OnChange(true)

	bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
	return string(bytes)

}

func DebugRequestLobbyStart(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}

	var args struct {
		Id *uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	lobby, _ := models.GetLobbyById(*args.Id)
	bytes, _ := json.Marshal(models.DecorateLobbyConnect(lobby))
	room := fmt.Sprintf("%s_private", chelpers.GetLobbyRoom(lobby.ID))
	broadcaster.SendMessageToRoom(room,
		"lobbyStart", string(bytes))

	bytes, _ = chelpers.BuildSuccessJSON(simplejson.New()).Encode()
	return string(bytes)
}

func DebugUpdateStatsFilter(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}

	var args struct {
		Id *uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	lobby, tperr := models.GetLobbyById(*args.Id)
	if tperr != nil {
		bytes, _ := tperr.ErrorJSON().Encode()
		return string(bytes)
	}
	lobby.UpdateStats()

	bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
	return string(bytes)
}

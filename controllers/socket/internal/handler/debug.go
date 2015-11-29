// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
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

func DebugPlayerSub(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}

	var args struct {
		Id    *uint   `json:"id"`
		Team  *string `json:"team"`
		Class *string `json:"class"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	lob, tperr := models.GetLobbyById(*args.Id)
	if tperr != nil {
		return tperr.Encode()
	}

	s, tperr := models.LobbyGetPlayerSlot(lob.Type, *args.Team, *args.Class)
	if tperr != nil {
		return tperr.Encode()
	}

	slot := models.LobbySlot{}
	err = db.DB.Where("slot = ?", s).First(&slot).Error
	if err != nil {
		helpers.Logger.Debug("", slot, s)
		return helpers.NewTPErrorFromError(err).Encode()
	}

	player := models.Player{}
	err = db.DB.First(&player, slot.PlayerId).Error
	if err != nil {
		helpers.Logger.Debug("", player)
		return helpers.NewTPErrorFromError(err).Encode()
	}

	sub, _ := models.NewSub(*args.Id, player.SteamId)
	db.DB.Save(sub)

	models.BroadcastSubList()

	return chelpers.EmptySuccessJS
}

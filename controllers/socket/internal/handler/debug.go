// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"fmt"
	"strconv"
	"time"

	"encoding/json"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/vibhavp/wsevent"
)

func DebugLobbyFill(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}
	var args struct {
		Id uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	lobby, _ := models.GetLobbyById(args.Id)
	var players []*models.Player

	for i := 1; i < int(lobby.Type)*2; i++ {
		steamid := "DEBUG" + strconv.FormatUint(uint64(time.Now().Unix()), 10) + strconv.Itoa(i)

		player, _ := models.NewPlayer(steamid)
		player.Debug = true
		player.Save()
		players = append(players, player)
		lobby.AddPlayer(player, i)
	}

	lobby.State = models.LobbyStateReadyingUp
	lobby.Save()
	room := fmt.Sprintf("%s_public", chelpers.GetLobbyRoom(lobby.ID))
	broadcaster.SendMessageToRoom(room, "lobbyReadyUp", "")
	lobby.ReadyUpTimeoutCheck()
	bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
	return string(bytes)

}

func DebugLobbyReady(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}

	var args struct {
		Id uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	lobby, _ := models.GetLobbyById(args.Id)

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

func DebugRequestAllLobbies(server *wsevent.Server, so *wsevent.Client, data string) string {
	var lobbies []models.Lobby
	db.DB.Where("state <> ?", models.LobbyStateEnded).Find(&lobbies)
	list, err := models.DecorateLobbyListData(lobbies)

	if err != nil {
		helpers.Logger.Warning("Failed to send lobby list: %s", err.Error())
	} else {
		reply, _ := json.Marshal(struct {
			Request string          `json:"request"`
			Data    json.RawMessage `json:"data"`
		}{"lobbyListData", []byte(list)})

		so.Emit(string(reply))
	}

	resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
	return string(resp)

}

func DebugRequestLobbyStart(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}

	var args struct {
		Id uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	lobby, _ := models.GetLobbyById(args.Id)
	bytes, _ := models.DecorateLobbyConnectJSON(lobby).Encode()
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
		Id uint `json:"id"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	lobby, tperr := models.GetLobbyById(args.Id)
	if err != nil {
		bytes, _ := tperr.ErrorJSON().Encode()
		return string(bytes)
	}
	lobby.UpdateStats()

	bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
	return string(bytes)
}

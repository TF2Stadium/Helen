// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

type Lobby struct{}

func (Lobby) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Lobby) LobbyCreate(_ *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		Map         *string `json:"map"`
		Type        *string `json:"type" valid:"debug,6s,highlander,4v4,ultiduo,bball"`
		League      *string `json:"league" valid:"ugc,etf2l,esea,asiafortress,ozfortress"`
		Server      *string `json:"server"`
		RconPwd     *string `json:"rconpwd"`
		WhitelistID *uint   `json:"whitelistID"`
		Mumble      *bool   `json:"mumbleRequired"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	var playermap = map[string]models.LobbyType{
		"debug":      models.LobbyTypeDebug,
		"6s":         models.LobbyTypeSixes,
		"highlander": models.LobbyTypeHighlander,
		"ultiduo":    models.LobbyTypeUltiduo,
		"bball":      models.LobbyTypeBball,
		"4v4":        models.LobbyTypeFours,
	}

	lobbyType := playermap[*args.Type]

	var count int
	db.DB.Table("server_records").Where("host = ?", *args.Server).Count(&count)
	if count != 0 {
		return helpers.NewTPError("A lobby is already using this server.", -1)
	}

	randBytes := make([]byte, 6)
	rand.Read(randBytes)
	serverPwd := base64.URLEncoding.EncodeToString(randBytes)

	//TODO what if playermap[lobbytype] is nil?
	info := models.ServerRecord{
		Host:           *args.Server,
		RconPassword:   *args.RconPwd,
		ServerPassword: serverPwd}
	// err = models.VerifyInfo(info)
	// if err != nil {
	// 	bytes, _ := helpers.NewTPErrorFromError(err).Encode()
	// 	return string(bytes)
	// }

	lob := models.NewLobby(*args.Map, lobbyType, *args.League, info, int(*args.WhitelistID), *args.Mumble)
	lob.CreatedBySteamID = player.SteamId
	lob.RegionCode, lob.RegionName = chelpers.GetRegion(*args.Server)
	if (lob.RegionCode == "" || lob.RegionName == "") && config.Constants.GeoIP != "" {
		return helpers.NewTPError("Couldn't find region server.", 1)
	}
	lob.Save()

	err = lob.SetupServer()
	if err != nil {
		qerr := db.DB.Where("id = ?", lob.ID).Delete(&models.Lobby{}).Error
		if qerr != nil {
			helpers.Logger.Warning(qerr.Error())
		}
		db.DB.Delete(&lob.ServerInfo)
		return helpers.NewTPErrorFromError(err)
	}

	lob.State = models.LobbyStateWaiting
	lob.Save()

	models.FumbleLobbyCreated(lob)

	return chelpers.BuildSuccessJSON(
		struct {
			ID uint `json:"id"`
		}{lob.ID})
}

func (Lobby) LobbyServerReset(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		ID *uint `json:"id"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
	if tperr != nil {
		return tperr
	}

	lobby, tperr := models.GetLobbyById(*args.ID)

	if player.SteamId != lobby.CreatedBySteamID || player.Role != helpers.RoleAdmin {
		return helpers.NewTPError("Player not authorized to reset server.", -1)
	}

	if tperr != nil {
		return tperr
	}

	if lobby.State == models.LobbyStateEnded {
		return helpers.NewTPError("Lobby has ended", 1)
	}

	if err := models.ReExecConfig(lobby.ID); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	return chelpers.EmptySuccessJS

}

func (Lobby) ServerVerify(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		Server  *string `json:"server"`
		Rconpwd *string `json:"rconpwd"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	var count int
	db.DB.Table("server_records").Where("host = ?", *args.Server).Count(&count)
	if count != 0 {
		return helpers.NewTPError("A lobby is already using this server.", -1)
	}

	info := &models.ServerRecord{
		Host:         *args.Server,
		RconPassword: *args.Rconpwd,
	}
	db.DB.Save(info)
	defer db.DB.Where("host = ?", info.Host).Delete(models.ServerRecord{})

	err := models.VerifyInfo(*info)
	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbyClose(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		Id *uint `json:"id"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)

	}

	player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	lob, tperr := models.GetLobbyByIdServer(uint(*args.Id))
	if tperr != nil {
		return tperr
	}

	if player.SteamId != lob.CreatedBySteamID && player.Role != helpers.RoleAdmin {
		return helpers.NewTPError("Player not authorized to close lobby.", -1)

	}

	if lob.State == models.LobbyStateEnded {
		return helpers.NewTPError("Lobby already closed.", -1)
	}

	models.FumbleLobbyEnded(lob)

	lob.Close(true)
	models.BroadcastLobbyList() // has to be done manually for now

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbyJoin(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr

	}

	var args struct {
		Id    *uint   `json:"id"`
		Class *string `json:"class"`
		Team  *string `json:"team" valid:"red,blu"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}
	//helpers.Logger.Debug("id %d class %s team %s", *args.Id, *args.Class, *args.Team)

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	if tperr != nil {
		return tperr
	}

	lob, tperr := models.GetLobbyById(*args.Id)
	if tperr != nil {
		return tperr
	}

	if lob.State == models.LobbyStateEnded {
		return helpers.NewTPError("Cannot join a closed lobby.", -1)

	}

	//Check if player is in the same lobby
	var sameLobby bool
	if id, err := player.GetLobbyId(); err == nil && id == *args.Id {
		sameLobby = true
	}

	slot, tperr := models.LobbyGetPlayerSlot(lob.Type, *args.Team, *args.Class)
	if tperr != nil {
		return tperr
	}

	if prevId, _ := player.GetLobbyId(); prevId != 0 && !sameLobby {
		server.RemoveClient(so.Id(), fmt.Sprintf("%d_public", prevId))
		server.RemoveClient(so.Id(), fmt.Sprintf("%d_private", prevId))
	}

	tperr = lob.AddPlayer(player, slot, *args.Team, *args.Class)

	if tperr != nil {
		return tperr
	}

	if !sameLobby {
		chelpers.AfterLobbyJoin(server, so, lob, player)
	}

	if lob.IsFull() {
		lob.State = models.LobbyStateReadyingUp
		lob.ReadyUpTimestamp = time.Now().Unix() + 30
		lob.Save()

		tick := time.After(time.Second * 30)
		id := lob.ID
		stop := make(chan struct{})

		go func() {
			select {
			case <-tick:
				lobby := &models.Lobby{}
				db.DB.First(lobby, id)

				if lobby.State != models.LobbyStateInProgress {
					err := lobby.RemoveUnreadyPlayers()
					if err != nil {
						helpers.Logger.Error("RemoveUnreadyPlayers: ", err.Error())
						err = nil
					}

					err = lobby.UnreadyAllPlayers()
					if err != nil {
						helpers.Logger.Error("UnreadyAllPlayers: ", err.Error())
					}

					lobby.State = models.LobbyStateWaiting
					lobby.Save()
				}

			case <-stop:
				return
			}
		}()

		room := fmt.Sprintf("%s_private",
			chelpers.GetLobbyRoom(lob.ID))
		broadcaster.SendMessageToRoom(room, "lobbyReadyUp",
			struct {
				Timeout int `json:"timeout"`
			}{30})
		models.BroadcastLobbyList()
	}

	err := models.AllowPlayer(*args.Id, player.SteamId, *args.Team+*args.Class)
	if err != nil {
		helpers.Logger.Error(err.Error())
	}

	if lob.State == models.LobbyStateInProgress {
		broadcaster.SendMessage(player.SteamId, "lobbyStart", models.DecorateLobbyConnect(lob, player.Name, *args.Class))
	}

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbySpectatorJoin(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		Id *uint `json:"id"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	var lob *models.Lobby
	lob, tperr := models.GetLobbyById(*args.Id)

	if tperr != nil {
		return tperr
	}

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
	if tperr != nil {
		return tperr
	}

	var specSameLobby bool

	arr, tperr := player.GetSpectatingIds()
	if len(arr) != 0 {
		for _, id := range arr {
			if id == *args.Id {
				specSameLobby = true
				continue
			}

			lobby, _ := models.GetLobbyById(id)
			lobby.RemoveSpectator(player, true)

			server.RemoveClient(so.Id(), fmt.Sprintf("%d_public", id))
		}
	}

	// If the player is already in the lobby (either joined a slot or is spectating), don't add them.
	// Just Broadcast the lobby to them, so the frontend displays it.
	if id, _ := player.GetLobbyId(); id != *args.Id && !specSameLobby {
		tperr = lob.AddSpectator(player)

		if tperr != nil {
			return tperr
		}
	}

	chelpers.AfterLobbySpec(server, so, lob)
	models.BroadcastLobbyToUser(lob, player.SteamId)
	return chelpers.EmptySuccessJS
}

func removePlayerFromLobby(lobbyId uint, steamId string) (*models.Lobby, *models.Player, *helpers.TPError) {
	player, tperr := models.GetPlayerBySteamId(steamId)
	if tperr != nil {
		return nil, nil, tperr
	}

	lob, tperr := models.GetLobbyById(lobbyId)
	if tperr != nil {
		return nil, nil, tperr
	}

	switch lob.State {
	case models.LobbyStateInProgress:
		return lob, player, helpers.NewTPError("Lobby is in progress.", 1)
	case models.LobbyStateEnded:
		return lob, player, helpers.NewTPError("Lobby has closed.", 1)
	}

	_, err := lob.GetPlayerSlot(player)
	if err != nil {
		return lob, player, helpers.NewTPError("Player not playing", 2)
	}

	if err := lob.RemovePlayer(player); err != nil {
		return lob, player, err
	}

	return lob, player, lob.AddSpectator(player)
}

func playerCanKick(lobbyId uint, steamId string) (bool, *helpers.TPError) {
	lob, tperr := models.GetLobbyById(lobbyId)
	if tperr != nil {
		return false, tperr
	}

	player, tperr2 := models.GetPlayerBySteamId(steamId)
	if tperr2 != nil {
		return false, tperr2
	}
	if steamId != lob.CreatedBySteamID && player.Role != helpers.RoleAdmin {
		return false, helpers.NewTPError("Not authorized to kick players", 1)
	}
	return true, nil
}

func (Lobby) LobbyKick(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		Id      *uint   `json:"id"`
		Steamid *string `json:"steamid"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	steamId := *args.Steamid
	selfSteamId := chelpers.GetSteamId(so.Id())

	if steamId == selfSteamId {
		return helpers.NewTPError("Player can't kick himself.", -1)
	}
	if ok, tperr := playerCanKick(*args.Id, selfSteamId); !ok {
		return tperr
	}

	lob, player, tperr := removePlayerFromLobby(*args.Id, steamId)
	if tperr != nil {
		return tperr
	}

	so, _ = broadcaster.GetSocket(player.SteamId)
	chelpers.AfterLobbyLeave(server, so, lob, player)

	// broadcaster.SendMessage(steamId, "sendNotification",
	// 	fmt.Sprintf(`{"notification": "You have been removed from Lobby #%d"}`, *args.Id))

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbyBan(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		Id      *uint   `json:"id"`
		Steamid *string `json:"steamid"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	steamId := *args.Steamid
	selfSteamId := chelpers.GetSteamId(so.Id())

	if steamId == selfSteamId {
		return helpers.NewTPError("Player can't kick himself.", -1)
	}
	if ok, tperr := playerCanKick(*args.Id, selfSteamId); !ok {
		return tperr
	}

	lob, player, tperr := removePlayerFromLobby(*args.Id, steamId)
	if tperr != nil {
		return tperr
	}

	lob.BanPlayer(player)

	so, _ = broadcaster.GetSocket(player.SteamId)
	chelpers.AfterLobbyLeave(server, so, lob, player)

	// broadcaster.SendMessage(steamId, "sendNotification",
	// 	fmt.Sprintf(`{"notification": "You have been removed from Lobby #%d"}`, *args.Id))

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbyLeave(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		Id *uint `json:"id"`
	}
	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	steamId := chelpers.GetSteamId(so.Id())

	lob, player, tperr := removePlayerFromLobby(*args.Id, steamId)
	if tperr != nil {
		return tperr
	}

	chelpers.AfterLobbyLeave(server, so, lob, player)

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbySpectatorLeave(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr
	}

	var args struct {
		Id *uint `json:"id"`
	}
	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	steamId := chelpers.GetSteamId(so.Id())
	player, tperr := models.GetPlayerBySteamId(steamId)
	if tperr != nil {
		return tperr
	}

	lob, tperr := models.GetLobbyById(*args.Id)
	if tperr != nil {
		return tperr
	}

	if !player.IsSpectatingId(lob.ID) {
		if id, _ := player.GetLobbyId(); id == *args.Id {
			chelpers.AfterLobbySpecLeave(server, so, lob)
			return chelpers.EmptySuccessJS
		}

		return helpers.NewTPError("Player is not spectating", -1)
	}

	lob.RemoveSpectator(player, true)
	chelpers.AfterLobbySpecLeave(server, so, lob)

	return chelpers.EmptySuccessJS
}

func (Lobby) RequestLobbyListData(_ *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	var lobbies []models.Lobby
	db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies)
	so.EmitJSON(helpers.NewRequest("lobbyListData", models.DecorateLobbyListData(lobbies)))

	return chelpers.EmptySuccessJS
}

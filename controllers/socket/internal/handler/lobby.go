// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

func LobbyCreate(_ *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
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
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
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
	lob.Region = chelpers.GetRegion(*args.Server)

	err = lob.SetupServer()
	if err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	lob.State = models.LobbyStateWaiting
	lob.Save()

	reply_str := struct {
		ID uint `json:"id"`
	}{lob.ID}
	bytes, _ := chelpers.BuildSuccessJSON(reply_str).Encode()
	return string(bytes)
}

func ServerVerify(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}

	var args struct {
		Server  *string `json:"server"`
		Rconpwd *string `json:"rconpwd"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	info := models.ServerRecord{
		Host:         *args.Server,
		RconPassword: *args.Rconpwd,
	}
	err := models.VerifyInfo(info)
	if err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	return chelpers.EmptySuccessJS
}

var timeoutStop = make(map[uint](chan bool))

func LobbyClose(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}

	var args struct {
		Id *uint `json:"id"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	lob, tperr := models.GetLobbyById(uint(*args.Id))
	if tperr != nil {
		bytes, _ := tperr.Encode()
		return string(bytes)
	}

	if player.SteamId != lob.CreatedBySteamID && player.Role != helpers.RoleAdmin {
		bytes, _ := helpers.NewTPError("Player not authorized to close lobby.", -1).Encode()
		return string(bytes)
	}

	if lob.State == models.LobbyStateEnded {
		bytes, _ := helpers.NewTPError("Lobby already closed.", -1).Encode()
		return string(bytes)
	}

	helpers.LockRecord(lob.ID, lob)
	lob.Close(true)
	helpers.UnlockRecord(lob.ID, lob)
	models.BroadcastLobbyList() // has to be done manually for now

	c, ok := timeoutStop[*args.Id]
	if !ok {
		c <- true
	}

	return chelpers.EmptySuccessJS
}

func LobbyJoin(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}

	var args struct {
		Id    *uint   `json:"id"`
		Class *string `json:"class"`
		Team  *string `json:"team" valid:"red,blu"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}
	//helpers.Logger.Debug("id %d class %s team %s", *args.Id, *args.Class, *args.Team)

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	lob, tperr := models.GetLobbyById(*args.Id)
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	if lob.State == models.LobbyStateEnded {
		bytes, _ := helpers.NewTPError("Cannot join a closed lobby.", -1).Encode()
		return string(bytes)
	}

	//Check if player is in the same lobby
	var sameLobby bool
	if id, err := player.GetLobbyId(); err == nil && id == *args.Id {
		sameLobby = true
	}

	slot, tperr := models.LobbyGetPlayerSlot(lob.Type, *args.Team, *args.Class)
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	helpers.LockRecord(lob.ID, lob)
	tperr = lob.AddPlayer(player, slot)
	helpers.UnlockRecord(lob.ID, lob)

	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
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
		timeoutStop[id] = make(chan bool)

		go func() {
			select {
			case <-tick:
				lobby := &models.Lobby{}
				db.DB.First(lobby, id)

				helpers.RLockRecord(id, lobby)
				state := lobby.State
				helpers.RUnlockRecord(id, lobby)

				if state != models.LobbyStateInProgress {
					helpers.LockRecord(lobby.ID, lobby)
					defer helpers.UnlockRecord(lobby.ID, lobby)
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

			case <-timeoutStop[id]:
				return
			}
		}()

		room := fmt.Sprintf("%s_private",
			chelpers.GetLobbyRoom(lob.ID))
		broadcaster.SendMessageToRoom(room, "lobbyReadyUp",
			`{"timeout":30}`)
		models.BroadcastLobbyList()
	}

	pub := fmt.Sprintf("%d_public", *args.Id)
	priv := fmt.Sprintf("%d_private", *args.Id)

	for _, room := range server.RoomsJoined(so.Id()) {
		if room == "0_public" {
			continue
		}
		if room != pub || room != priv {
			server.RemoveClient(so.Id(), room)
		}
	}
	models.AllowPlayer(*args.Id, player.SteamId, *args.Team+*args.Class)
	models.BroadcastLobbyToUser(lob, player.SteamId)

	return chelpers.EmptySuccessJS
}

func LobbySpectatorJoin(server *wsevent.Server, so *wsevent.Client, data string) string {
	var noLogin bool
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		noLogin = true
		// bytes, _ := json.Marshal(reqerr)
		// return string(bytes)
	}

	var args struct {
		Id *uint `json:"id"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	var lob *models.Lobby
	lob, tperr := models.GetLobbyById(*args.Id)

	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	if noLogin {
		chelpers.AfterLobbySpec(server, so, lob)
		bytes, _ := json.Marshal(models.DecorateLobbyData(lob, true))

		so.EmitJSON(helpers.NewRequest("lobbyData", string(bytes)))

		return chelpers.EmptySuccessJS
	}

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	if id, _ := player.GetLobbyId(); id != *args.Id {
		helpers.LockRecord(lob.ID, lob)
		tperr = lob.AddSpectator(player)
		helpers.UnlockRecord(lob.ID, lob)
	}

	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	chelpers.AfterLobbySpec(server, so, lob)
	models.BroadcastLobbyToUser(lob, player.SteamId)
	return chelpers.EmptySuccessJS
}

func LobbyKick(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}

	var args struct {
		Id      *uint   `json:"id"`
		Steamid *string `json:"steamid"`
		Ban     *bool   `json:"ban" empty:"false"`
	}

	if err := chelpers.GetParams(data, &args); err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	steamid := *args.Steamid
	var self bool

	selfSteamid := chelpers.GetSteamId(so.Id())
	// TODO check authorization, currently can kick anyone

	if steamid == "" || steamid == selfSteamid {
		self = true
		steamid = selfSteamid
	}

	if self && *args.Ban {
		bytes, _ := helpers.NewTPError("Player can't ban himself.", -1).Encode()
		return string(bytes)
	}

	//player to kick
	player, tperr := models.GetPlayerBySteamId(steamid)
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	playerRequesting, tperr2 := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
	if tperr2 != nil {
		bytes, _ := json.Marshal(tperr2)
		return string(bytes)
	}

	lob, tperr := models.GetLobbyById(*args.Id)
	if tperr != nil {
		bytes, _ := tperr.Encode()
		return string(bytes)
	}

	switch lob.State {
	case models.LobbyStateInProgress:
		bytes, _ := helpers.NewTPError("Lobby is in progress.", 1).Encode()
		return string(bytes)
	case models.LobbyStateEnded:
		bytes, _ := helpers.NewTPError("Lobby has closed.", 1).Encode()
		return string(bytes)
	}

	if !self && selfSteamid != lob.CreatedBySteamID && playerRequesting.Role != helpers.RoleAdmin {
		bytes, _ := helpers.NewTPError(
			"Not authorized to kick players", 1).Encode()
		return string(bytes)
	}

	_, err := lob.GetPlayerSlot(player)
	helpers.LockRecord(lob.ID, lob)
	defer helpers.UnlockRecord(lob.ID, lob)

	var spec bool
	if err == nil {
		lob.RemovePlayer(player)
	} else if player.IsSpectatingId(lob.ID) {
		spec = true
		lob.RemoveSpectator(player, true)
	} else {
		bytes, _ := helpers.NewTPError("Player neither playing nor spectating", 2).Encode()
		return string(bytes)
	}

	if *args.Ban {
		fmt.Println(playerRequesting.Role)
		if playerRequesting.Role == helpers.RoleAdmin {
			lob.BanPlayer(player)
		} else {
			bytes, _ := helpers.NewTPError(
				"Not authorized to ban players", 1).Encode()
			return string(bytes)
		}
	}

	if !self {
		so, _ = broadcaster.GetSocket(player.SteamId)
	}

	if !spec {
		chelpers.AfterLobbyLeave(server, so, lob, player)
	} else {
		chelpers.AfterLobbySpecLeave(server, so, lob)
	}

	if !self {
		broadcaster.SendMessage(steamid, "sendNotification",
			fmt.Sprintf(`{"notification": "You have been removed from Lobby #%d"}`,
				*args.Id))

	}

	return chelpers.EmptySuccessJS
}

func PlayerReady(_ *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}

	steamid := chelpers.GetSteamId(so.Id())
	player, tperr := models.GetPlayerBySteamId(steamid)
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	lobbyid, tperr := player.GetLobbyId()
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	lobby, tperr := models.GetLobbyById(lobbyid)
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	if lobby.State != models.LobbyStateReadyingUp {
		bytes, _ := json.Marshal(helpers.NewTPError("Lobby hasn't been filled up yet.", 4))
		return string(bytes)
	}

	helpers.LockRecord(lobby.ID, lobby)
	tperr = lobby.ReadyPlayer(player)
	defer helpers.UnlockRecord(lobby.ID, lobby)

	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	if lobby.IsEveryoneReady() {
		timeoutStop[lobby.ID] <- true
		lobby.State = models.LobbyStateInProgress
		lobby.Save()
		bytes, _ := json.Marshal(models.DecorateLobbyConnect(lobby))
		room := fmt.Sprintf("%s_private",
			chelpers.GetLobbyRoom(lobby.ID))
		broadcaster.SendMessageToRoom(room,
			"lobbyStart", string(bytes))
		models.BroadcastLobbyList()
	}

	return chelpers.EmptySuccessJS
}

func PlayerNotReady(_ *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	lobbyid, tperr := player.GetLobbyId()
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	lobby, tperr := models.GetLobbyById(lobbyid)
	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	if lobby.State != models.LobbyStateReadyingUp {
		bytes, _ := json.Marshal(helpers.NewTPError("Lobby hasn't been filled up yet.", 4))
		return string(bytes)
	}

	helpers.LockRecord(lobby.ID, lobby)
	tperr = lobby.UnreadyPlayer(player)
	lobby.RemovePlayer(player)
	helpers.UnlockRecord(lobby.ID, lobby)

	if tperr != nil {
		bytes, _ := json.Marshal(tperr)
		return string(bytes)
	}

	lobby.UnreadyAllPlayers()
	c, ok := timeoutStop[lobby.ID]
	if ok {
		c <- true
	}

	return chelpers.EmptySuccessJS
}

func RequestLobbyListData(_ *wsevent.Server, so *wsevent.Client, data string) string {
	var lobbies []models.Lobby
	db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies)
	list, _ := json.Marshal(models.DecorateLobbyListData(lobbies))
	so.EmitJSON(helpers.NewRequest("lobbyListData", string(list)))

	return chelpers.EmptySuccessJS
}

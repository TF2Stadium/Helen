// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

type Lobby struct{}

func (Lobby) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

var rSteamGroup = regexp.MustCompile(`steamcommunity\.com\/groups\/(.+)`)

type Restriction struct {
	Red bool `json:"red,omitempty"`
	Blu bool `json:"blu,omitempty"`
}
type Requirement struct {
	Hours      int         `json:"hours"`
	Lobbies    int         `json:"lobbies"`
	Restricted Restriction `json:"restricted"`
}

func newRequirement(team, class string, requirement Requirement, lobby *models.Lobby) *helpers.TPError {
	slot, err := models.LobbyGetPlayerSlot(lobby.Type, team, class)
	if err != nil {
		return err
	}
	slotReq := &models.Requirement{
		LobbyID: lobby.ID,
		Slot:    slot,
		Hours:   requirement.Hours,
		Lobbies: requirement.Lobbies,
	}
	slotReq.Save()

	return nil
}

func (Lobby) LobbyCreate(so *wsevent.Client, args struct {
	Map         *string `json:"map"`
	Type        *string `json:"type" valid:"debug,6s,highlander,4v4,ultiduo,bball"`
	League      *string `json:"league" valid:"ugc,etf2l,esea,asiafortress,ozfortress"`
	Server      *string `json:"server"`
	RconPwd     *string `json:"rconpwd"`
	WhitelistID *string `json:"whitelistID"`
	Mumble      *bool   `json:"mumbleRequired"`

	Password            *string `json:"password" empty:"-"`
	SteamGroupWhitelist *string `json:"steamGroupWhitelist" empty:"-"`

	Requirements *struct {
		Classes map[string]Requirement `json:"classes,omitempty"`
		General Requirement            `json:"general,omitempty"`
	} `json:"requirements" empty:"-"`
}) interface{} {

	player := chelpers.GetPlayerFromSocket(so.ID)
	if banned, until := player.IsBannedWithTime(models.PlayerBanCreate); banned {
		str := fmt.Sprintf("You've been banned from creating lobbies till %s", until.Format(time.RFC822))
		return helpers.NewTPError(str, -1)
	}

	var steamGroup string
	if *args.SteamGroupWhitelist != "" && !rSteamGroup.MatchString(*args.SteamGroupWhitelist) {
		return helpers.NewTPError("Invalid Steam group URL", -1)
	} else if rSteamGroup.MatchString(*args.SteamGroupWhitelist) {
		steamGroup = rSteamGroup.FindStringSubmatch(*args.SteamGroupWhitelist)[1]
	}

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

	lob := models.NewLobby(*args.Map, lobbyType, *args.League, info, *args.WhitelistID, *args.Mumble, steamGroup, *args.Password)
	lob.CreatedBySteamID = player.SteamID
	lob.RegionCode, lob.RegionName = chelpers.GetRegion(*args.Server)
	if (lob.RegionCode == "" || lob.RegionName == "") && config.Constants.GeoIP != "" {
		return helpers.NewTPError("Couldn't find region server.", 1)
	}
	lob.Save()

	err := lob.SetupServer()
	if err != nil { //lobby setup failed, delete lobby and corresponding server record
		qerr := db.DB.Where("id = ?", lob.ID).Delete(&models.Lobby{}).Error
		if qerr != nil {
			logrus.Warning(qerr.Error())
		}
		db.DB.Delete(&lob.ServerInfo)
		return helpers.NewTPErrorFromError(err)
	}

	lob.State = models.LobbyStateWaiting
	lob.Save()

	models.FumbleLobbyCreated(lob)

	if args.Requirements != nil {
		for class, requirement := range (*args.Requirements).Classes {
			if requirement.Restricted.Blu {
				err := newRequirement("blu", class, requirement, lob)
				if err != nil {
					return err
				}
			}
			if requirement.Restricted.Red {
				err := newRequirement("red", class, requirement, lob)
				if err != nil {
					return err
				}
			}
		}
		if args.Requirements.General.Hours != 0 || args.Requirements.General.Lobbies != 0 {
			general := &models.Requirement{
				LobbyID: lob.ID,
				Hours:   args.Requirements.General.Hours,
				Lobbies: args.Requirements.General.Lobbies,
				Slot:    -1,
			}
			general.Save()
		}
	}
	return chelpers.NewResponse(
		struct {
			ID uint `json:"id"`
		}{lob.ID})
}

func (Lobby) LobbyServerReset(so *wsevent.Client, args struct {
	ID *uint `json:"id"`
}) interface{} {

	player := chelpers.GetPlayerFromSocket(so.ID)
	lobby, tperr := models.GetLobbyByID(*args.ID)

	if player.SteamID != lobby.CreatedBySteamID || player.Role != helpers.RoleAdmin {
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

var validAddress = regexp.MustCompile(`.+\:\d+`)

func (Lobby) ServerVerify(so *wsevent.Client, args struct {
	Server  *string `json:"server"`
	Rconpwd *string `json:"rconpwd"`
}) interface{} {

	if !validAddress.MatchString(*args.Server) {
		return helpers.NewTPError("Invalid Server Address", -1)
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

func (Lobby) LobbyClose(so *wsevent.Client, args struct {
	Id *uint `json:"id"`
}) interface{} {

	player := chelpers.GetPlayerFromSocket(so.ID)
	lob, tperr := models.GetLobbyByIDServer(uint(*args.Id))
	if tperr != nil {
		return tperr
	}

	if player.SteamID != lob.CreatedBySteamID && player.Role != helpers.RoleAdmin {
		return helpers.NewTPError("Player not authorized to close lobby.", -1)

	}

	if lob.State == models.LobbyStateEnded {
		return helpers.NewTPError("Lobby already closed.", -1)
	}

	lob.Close(true)

	notify := fmt.Sprintf("Lobby closed by %s", player.Name)
	models.SendNotification(notify, int(lob.ID))

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbyJoin(so *wsevent.Client, args struct {
	Id       *uint   `json:"id"`
	Class    *string `json:"class"`
	Team     *string `json:"team" valid:"red,blu"`
	Password *string `json:"password" empty:"-"`
}) interface{} {

	player := chelpers.GetPlayerFromSocket(so.ID)
	if banned, until := player.IsBannedWithTime(models.PlayerBanJoin); banned {
		str := fmt.Sprintf("You have been banned from joining lobbies till %s", until.Format(time.RFC822))
		return helpers.NewTPError(str, -1)
	}

	//logrus.Debug("id %d class %s team %s", *args.Id, *args.Class, *args.Team)
	lob, tperr := models.GetLobbyByID(*args.Id)
	if tperr != nil {
		return tperr
	}

	if lob.State == models.LobbyStateEnded {
		return helpers.NewTPError("Cannot join a closed lobby.", -1)

	}

	//Check if player is in the same lobby
	var sameLobby bool
	if id, err := player.GetLobbyID(false); err == nil && id == *args.Id {
		sameLobby = true
	}

	slot, tperr := models.LobbyGetPlayerSlot(lob.Type, *args.Team, *args.Class)
	if tperr != nil {
		return tperr
	}

	if prevId, _ := player.GetLobbyID(false); prevId != 0 && !sameLobby {
		lobby, _ := models.GetLobbyByID(prevId)
		hooks.AfterLobbyLeave(lobby, player)
	}

	tperr = lob.AddPlayer(player, slot, *args.Password)

	if tperr != nil {
		return tperr
	}

	if !sameLobby {
		hooks.AfterLobbyJoin(so, lob, player)
	}

	//check if lobby isn't already in progress (which happens when the player is subbing)
	if lob.IsFull() && lob.State != models.LobbyStateInProgress {
		lob.State = models.LobbyStateReadyingUp
		lob.ReadyUpTimestamp = time.Now().Unix() + 30
		lob.Save()

		time.AfterFunc(time.Second*30, func() {
			lobby := &models.Lobby{}
			db.DB.First(lobby, lob.ID)

			//if all player's haven't readied up,
			//remove unreadied players and unready the
			//rest.
			if lobby.State != models.LobbyStateInProgress && lobby.State != models.LobbyStateEnded {
				removeUnreadyPlayers(lobby)

				lobby.State = models.LobbyStateWaiting
				lobby.Save()
			}
		})

		room := fmt.Sprintf("%s_private",
			hooks.GetLobbyRoom(lob.ID))
		broadcaster.SendMessageToRoom(room, "lobbyReadyUp",
			struct {
				Timeout int `json:"timeout"`
			}{30})
		models.BroadcastLobbyList()
	}

	if lob.State == models.LobbyStateInProgress { //this happens when the player is a substitute
		db.DB.Preload("ServerInfo").First(lob, lob.ID)
		so.EmitJSON(helpers.NewRequest("lobbyStart", models.DecorateLobbyConnect(lob, player.Name, slot)))
	}

	return chelpers.EmptySuccessJS
}

func removeUnreadyPlayers(lobby *models.Lobby) {
	var players []*models.Player

	db.DB.Table("players").Joins("INNER JOIN lobby_slots ON lobby_slots.player_id = players.id").Where("lobby_slots.lobby_id = ? AND lobby_slots.ready = ?", lobby.ID, false).Find(&players)
	lobby.RemoveUnreadyPlayers(true)

	for _, player := range players {
		hooks.AfterLobbyLeave(lobby, player)
	}
}

func (Lobby) LobbySpectatorJoin(so *wsevent.Client, args struct {
	Id *uint `json:"id"`
}) interface{} {

	var lob *models.Lobby
	lob, tperr := models.GetLobbyByID(*args.Id)

	if tperr != nil {
		return tperr
	}

	player := chelpers.GetPlayerFromSocket(so.ID)
	var specSameLobby bool

	arr, tperr := player.GetSpectatingIds()
	if len(arr) != 0 {
		for _, id := range arr {
			if id == *args.Id {
				specSameLobby = true
				continue
			}
			//a socket should only spectate one lobby, remove socket from
			//any other lobby room
			//multiple sockets from one player can spectatte multiple lobbies
			socket.AuthServer.Leave(so, fmt.Sprintf("%d_public", id))
		}
	}

	// If the player is already in the lobby (either joined a slot or is spectating), don't add them.
	// Just Broadcast the lobby to them, so the frontend displays it.
	if id, _ := player.GetLobbyID(false); id != *args.Id && !specSameLobby {
		tperr = lob.AddSpectator(player)

		if tperr != nil {
			return tperr
		}
	}

	hooks.AfterLobbySpec(socket.AuthServer, so, lob)
	models.BroadcastLobbyToUser(lob, player.SteamID)
	return chelpers.EmptySuccessJS
}

func removePlayerFromLobby(lobbyId uint, steamId string) (*models.Lobby, *models.Player, *helpers.TPError) {
	player, tperr := models.GetPlayerBySteamID(steamId)
	if tperr != nil {
		return nil, nil, tperr
	}

	lob, tperr := models.GetLobbyByID(lobbyId)
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
	lob, tperr := models.GetLobbyByID(lobbyId)
	if tperr != nil {
		return false, tperr
	}

	player, tperr2 := models.GetPlayerBySteamID(steamId)
	if tperr2 != nil {
		return false, tperr2
	}
	if steamId != lob.CreatedBySteamID && player.Role != helpers.RoleAdmin {
		return false, helpers.NewTPError("Not authorized to kick players", 1)
	}
	return true, nil
}

func (Lobby) LobbyKick(so *wsevent.Client, args struct {
	Id      *uint   `json:"id"`
	Steamid *string `json:"steamid"`
}) interface{} {

	steamId := *args.Steamid
	selfSteamId := chelpers.GetSteamId(so.ID)

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

	hooks.AfterLobbyLeave(lob, player)

	// broadcaster.SendMessage(steamId, "sendNotification",
	// 	fmt.Sprintf(`{"notification": "You have been removed from Lobby #%d"}`, *args.Id))

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbyBan(so *wsevent.Client, args struct {
	Id      *uint   `json:"id"`
	Steamid *string `json:"steamid"`
}) interface{} {

	steamId := *args.Steamid
	selfSteamId := chelpers.GetSteamId(so.ID)

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

	hooks.AfterLobbyLeave(lob, player)

	// broadcaster.SendMessage(steamId, "sendNotification",
	// 	fmt.Sprintf(`{"notification": "You have been removed from Lobby #%d"}`, *args.Id))

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbyLeave(so *wsevent.Client, args struct {
	Id *uint `json:"id"`
}) interface{} {

	steamId := chelpers.GetSteamId(so.ID)

	lob, player, tperr := removePlayerFromLobby(*args.Id, steamId)
	if tperr != nil {
		return tperr
	}

	hooks.AfterLobbyLeave(lob, player)

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbySpectatorLeave(so *wsevent.Client, args struct {
	Id *uint `json:"id"`
}) interface{} {

	player := chelpers.GetPlayerFromSocket(so.ID)
	lob, tperr := models.GetLobbyByID(*args.Id)
	if tperr != nil {
		return tperr
	}

	if !player.IsSpectatingID(lob.ID) {
		if id, _ := player.GetLobbyID(false); id == *args.Id {
			hooks.AfterLobbySpecLeave(so, lob)
			return chelpers.EmptySuccessJS
		}
	}

	lob.RemoveSpectator(player, true)
	hooks.AfterLobbySpecLeave(so, lob)

	return chelpers.EmptySuccessJS
}

func (Lobby) RequestLobbyListData(so *wsevent.Client, _ struct{}) interface{} {
	var lobbies []models.Lobby
	db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies)
	so.EmitJSON(helpers.NewRequest("lobbyListData", models.DecorateLobbyListData(lobbies)))

	return chelpers.EmptySuccessJS
}

func (Lobby) LobbyChangeOwner(so *wsevent.Client, args struct {
	ID      *uint   `json:"id"`
	SteamID *string `json:"steamid"`
}) interface{} {
	lobby, err := models.GetLobbyByID(*args.ID)
	if err != nil {
		return err
	}

	player := chelpers.GetPlayerFromSocket(so.ID)
	if lobby.CreatedBySteamID != player.SteamID {
		return helpers.NewTPError("You aren't authorized to change lobby owner.", -1)
	}

	player2, err := models.GetPlayerBySteamID(*args.SteamID)
	if err != nil {
		return err
	}

	lobby.CreatedBySteamID = player2.SteamID
	lobby.Save()
	models.BroadcastLobby(lobby)
	models.BroadcastLobbyList()
	models.NewBotMessage(fmt.Sprintf("Lobby leader changed to %s", player.Name), int(*args.ID)).Send()

	return chelpers.EmptySuccessJS
}

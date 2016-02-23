// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

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

var (
	reSteamGroup = regexp.MustCompile(`steamcommunity\.com\/groups\/(.+)`)
	reTwitchChan = regexp.MustCompile(`twitch.tv\/(.+)`)
)

type Restriction struct {
	Red bool `json:"red,omitempty"`
	Blu bool `json:"blu,omitempty"`
}
type Requirement struct {
	Hours      int         `json:"hours"`
	Lobbies    int         `json:"lobbies"`
	Restricted Restriction `json:"restricted"`
}

func newRequirement(team, class string, requirement Requirement, lobby *models.Lobby) error {
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
	// restrict lobby slots to twitch subs for a particular channel
	TwitchWhitelist *string `json:"twitchWhitelist" empty:"-"`

	Requirements *struct {
		Classes map[string]Requirement `json:"classes,omitempty"`
		General Requirement            `json:"general,omitempty"`
	} `json:"requirements" empty:"-"`
}) interface{} {

	player := chelpers.GetPlayerFromSocket(so.ID)
	if banned, until := player.IsBannedWithTime(models.PlayerBanCreate); banned {
		return fmt.Errorf("You've been banned from creating lobbies till %s", until.Format(time.RFC822))
	}

	var steamGroup string
	var twitchChan string
	if *args.SteamGroupWhitelist != "" {
		if reSteamGroup.MatchString(*args.SteamGroupWhitelist) {
			steamGroup = reSteamGroup.FindStringSubmatch(*args.SteamGroupWhitelist)[1]
		} else {
			return errors.New("Invalid Steam group URL")
		}
	}
	if *args.TwitchWhitelist != "" {
		if reTwitchChan.MatchString(*args.TwitchWhitelist) {
			twitchChan = reTwitchChan.FindStringSubmatch(*args.TwitchWhitelist)[1]
		} else {
			return errors.New("Invalid Twitch channel URL")
		}
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
		return errors.New("A lobby is already using this server.")
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
	lob.TwitchChannel = twitchChan
	lob.CreatedBySteamID = player.SteamID
	lob.RegionCode, lob.RegionName = chelpers.GetRegion(*args.Server)
	if (lob.RegionCode == "" || lob.RegionName == "") && config.Constants.GeoIP {
		return errors.New("Couldn't find region server.")
	}
	lob.Save()

	err := lob.SetupServer()
	if err != nil { //lobby setup failed, delete lobby and corresponding server record
		lob.Delete()
		return err
	}

	lob.State = models.LobbyStateWaiting
	lob.Save()

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
			for i := 0; i < 2*models.NumberOfClassesMap[lob.Type]; i++ {
				req := &models.Requirement{
					LobbyID: lob.ID,
					Hours:   args.Requirements.General.Hours,
					Lobbies: args.Requirements.General.Lobbies,
					Slot:    i,
				}
				req.Save()
			}

		}
	}
	return newResponse(
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
		return errors.New("Player not authorized to reset server.")
	}

	if tperr != nil {
		return tperr
	}

	if lobby.State == models.LobbyStateEnded {
		return errors.New("Lobby has ended")
	}

	if err := models.ReExecConfig(lobby.ID); err != nil {
		return err
	}

	return emptySuccess
}

var validAddress = regexp.MustCompile(`.+\:\d+`)

func (Lobby) ServerVerify(so *wsevent.Client, args struct {
	Server  *string `json:"server"`
	Rconpwd *string `json:"rconpwd"`
}) interface{} {

	if !validAddress.MatchString(*args.Server) {
		return errors.New("Invalid Server Address")
	}

	var count int
	db.DB.Table("server_records").Where("host = ?", *args.Server).Count(&count)
	if count != 0 {
		return errors.New("A lobby is already using this server.")
	}

	info := &models.ServerRecord{
		Host:         *args.Server,
		RconPassword: *args.Rconpwd,
	}
	db.DB.Save(info)
	defer db.DB.Delete(info)

	err := models.VerifyInfo(*info)
	if err != nil {
		return err
	}

	return emptySuccess
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
		return errors.New("Player not authorized to close lobby.")

	}

	if lob.State == models.LobbyStateEnded {
		return errors.New("Lobby already closed.")
	}

	lob.Close(true)

	notify := fmt.Sprintf("Lobby closed by %s", player.Alias())
	models.SendNotification(notify, int(lob.ID))

	return emptySuccess
}

func (Lobby) LobbyJoin(so *wsevent.Client, args struct {
	Id       *uint   `json:"id"`
	Class    *string `json:"class"`
	Team     *string `json:"team" valid:"red,blu"`
	Password *string `json:"password" empty:"-"`
}) interface{} {

	player := chelpers.GetPlayerFromSocket(so.ID)
	if banned, until := player.IsBannedWithTime(models.PlayerBanJoin); banned {
		return fmt.Errorf("You have been banned from joining lobbies till %s", until.Format(time.RFC822))
	}

	//logrus.Debug("id %d class %s team %s", *args.Id, *args.Class, *args.Team)
	lob, tperr := models.GetLobbyByID(*args.Id)
	if tperr != nil {
		return tperr
	}

	if lob.State == models.LobbyStateEnded {
		return errors.New("Cannot join a closed lobby.")

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
			state := lob.CurrentState()
			//if all player's haven't readied up,
			//remove unreadied players and unready the
			//rest.
			if state != models.LobbyStateInProgress && state != models.LobbyStateEnded {
				removeUnreadyPlayers(lob)
				lob.SetState(models.LobbyStateWaiting)
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

	return emptySuccess
}

//get list of unready players, remove them from lobby (and add them as spectators)
//plus, call the after lobby leave hook for each player removed
func removeUnreadyPlayers(lobby *models.Lobby) {
	players := lobby.GetUnreadyPlayers()
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
	return emptySuccess
}

func removePlayerFromLobby(lobbyId uint, steamId string) (*models.Lobby, *models.Player, error) {
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
		return lob, player, errors.New("Lobby is in progress.")
	case models.LobbyStateEnded:
		return lob, player, errors.New("Lobby has closed.")
	}

	_, err := lob.GetPlayerSlot(player)
	if err != nil {
		return lob, player, errors.New("Player not playing")
	}

	if err := lob.RemovePlayer(player); err != nil {
		return lob, player, err
	}

	return lob, player, lob.AddSpectator(player)
}

func playerCanKick(lobbyId uint, steamId string) (bool, error) {
	lob, tperr := models.GetLobbyByID(lobbyId)
	if tperr != nil {
		return false, tperr
	}

	player, tperr2 := models.GetPlayerBySteamID(steamId)
	if tperr2 != nil {
		return false, tperr2
	}
	if steamId != lob.CreatedBySteamID && player.Role != helpers.RoleAdmin {
		return false, errors.New("Not authorized to kick players")
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
		return errors.New("Player can't kick himself.")
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

	return emptySuccess
}

func (Lobby) LobbyBan(so *wsevent.Client, args struct {
	Id      *uint   `json:"id"`
	Steamid *string `json:"steamid"`
}) interface{} {

	steamId := *args.Steamid
	selfSteamId := chelpers.GetSteamId(so.ID)

	if steamId == selfSteamId {
		return errors.New("Player can't kick himself.")
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

	return emptySuccess
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

	return emptySuccess
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
			return emptySuccess
		}
	}

	lob.RemoveSpectator(player, true)
	hooks.AfterLobbySpecLeave(so, lob)

	return emptySuccess
}

func (Lobby) RequestLobbyListData(so *wsevent.Client, _ struct{}) interface{} {
	so.EmitJSON(helpers.NewRequest("lobbyListData", models.DecorateLobbyListData(models.GetWaitingLobbies())))

	return emptySuccess
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
		return errors.New("You aren't authorized to change lobby owner.")
	}

	player2, err := models.GetPlayerBySteamID(*args.SteamID)
	if err != nil {
		return err
	}

	lobby.CreatedBySteamID = player2.SteamID
	lobby.Save()
	models.BroadcastLobby(lobby)
	models.BroadcastLobbyList()
	models.NewBotMessage(fmt.Sprintf("Lobby leader changed to %s", player.Alias()), int(*args.ID)).Send()

	return emptySuccess
}

func (Lobby) LobbySetRequirement(so *wsevent.Client, args struct {
	ID *uint `json:"id"` // lobby ID

	Slot  *int         `json:"slot"` // -1 if to set for all slots
	Type  *string      `json:"type"`
	Value *json.Number `json:"value"`
}) interface{} {

	lobby, tperr := models.GetLobbyByID(*args.ID)
	if tperr != nil {
		return tperr
	}

	player := chelpers.GetPlayerFromSocket(so.ID)
	if lobby.CreatedBySteamID != player.SteamID {
		return errors.New("Only lobby owners can change requirements.")
	}

	if !(*args.Slot >= 0 && *args.Slot < models.NumberOfClassesMap[lobby.Type]) {
		return errors.New("Invalid slot.")
	}

	req, err := lobby.GetSlotRequirement(*args.Slot)
	if err != nil { //requirement doesn't exist. create one
		req.Slot = *args.Slot
		req.LobbyID = *args.ID
	}

	var n int64
	var f float64

	switch *args.Type {
	case "hours":
		n, err = (*args.Value).Int64()
		req.Hours = int(n)
	case "lobbies":
		n, err = (*args.Value).Int64()
		req.Lobbies = int(n)
	case "reliability":
		f, err = (*args.Value).Float64()
		req.Reliability = f
	default:
		return errors.New("Invalid requirement type.")
	}

	if err != nil {
		return errors.New("Invalid requirement.")
	}

	req.Save()
	models.BroadcastLobby(lobby)
	models.BroadcastLobbyList()

	return emptySuccess
}

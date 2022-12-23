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
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/chat"
	"github.com/TF2Stadium/Helen/models/gameserver"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/lobby/format"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/Helen/models/rpc"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/servemetf"
	"github.com/TF2Stadium/wsevent"
)

type Lobby struct{}

func (Lobby) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

var (
	reDiscordInvite = regexp.MustCompile(`https:\/\/discord.gg\/[a-zA-Z0-9]+`)
	reSteamGroup    = regexp.MustCompile(`steamcommunity\.com\/groups\/(.+)`)
	reServer        = regexp.MustCompile(`\w+\:\d+`)
	playermap       = map[string]format.Format{
		"debug":      format.Debug,
		"6s":         format.Sixes,
		"highlander": format.Highlander,
		"ultiduo":    format.Ultiduo,
		"bball":      format.Bball,
		"4v4":        format.Fours,
		"prolander":  format.Prolander,
	}
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

type servemeServer struct {
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
	Server   servemetf.Server
}

func newRequirement(team, class string, requirement Requirement, lob *lobby.Lobby) error {
	slot, err := format.GetSlot(lob.Type, team, class)
	if err != nil {
		return err
	}
	slotReq := &lobby.Requirement{
		LobbyID: lob.ID,
		Slot:    slot,
		Hours:   int(requirement.Hours),
		Lobbies: int(requirement.Lobbies),
	}
	slotReq.Save()

	return nil
}

func (Lobby) LobbyCreate(so *wsevent.Client, args struct {
	Map         *string        `json:"map"`
	Type        *string        `json:"type" valid:"debug,6s,highlander,4v4,ultiduo,bball,prolander"`
	League      *string        `json:"league" valid:"ugc,etf2l,esea,asiafortress,ozfortress,bballtf,rgl"`
	ServerType  *string        `json:"serverType" valid:"server,storedServer,serveme"`
	Serveme     *servemeServer `json:"serveme" empty:"-"`
	Server      *string        `json:"server" empty:"-"`
	RconPwd     *string        `json:"rconpwd" empty:"-"`
	WhitelistID *string        `json:"whitelistID"`
	Mumble      *bool          `json:"mumbleRequired"`

	Password            *string `json:"password" empty:"-"`
	SteamGroupWhitelist *string `json:"steamGroupWhitelist" empty:"-"`
	// restrict lobby slots to twitch subs for a particular channel
	// not a pointer, since it is set to false when the argument json
	// string doesn't have the field
	TwitchWhitelistSubscribers bool `json:"twitchWhitelistSubs"`
	TwitchWhitelistFollowers   bool `json:"twitchWhitelistFollows"`
	RegionLock                 bool `json:"regionLock"`

	Requirements *struct {
		Classes map[string]Requirement `json:"classes,omitempty"`
		General Requirement            `json:"general,omitempty"`
	} `json:"requirements" empty:"-"`

	Discord *struct {
		RedChannel *string `json:"redChannel,omitempty"`
		BluChannel *string `json:"bluChannel,omitempty"`
	} `json:"discord" empty:"-"`
}) interface{} {
	p := chelpers.GetPlayer(so.Token)
	if banned, until := p.IsBannedWithTime(player.BanCreate); banned {
		ban, _ := p.GetActiveBan(player.BanCreate)
		return fmt.Errorf("You've been banned from creating lobbies till %s (%s)", until.Format(time.RFC822), ban.Reason)
	}

	if p.HasCreatedLobby() {
		if p.Role != helpers.RoleAdmin && p.Role != helpers.RoleMod {
			return errors.New("You have already created a lobby.")
		}
	}

	var steamGroup string
	var context *servemetf.Context
	var reservation servemetf.Reservation

	if *args.SteamGroupWhitelist != "" {
		if reSteamGroup.MatchString(*args.SteamGroupWhitelist) {
			steamGroup = reSteamGroup.FindStringSubmatch(*args.SteamGroupWhitelist)[1]
		} else {
			return errors.New("Invalid Steam group URL")
		}
	}

	if *args.ServerType == "serveme" {
		if args.Serveme == nil {
			return errors.New("No serveme info given.")
		}
		var err error
		var start, end time.Time

		if start, err = time.Parse(servemetf.TimeFormat, (*args.Serveme).StartsAt); err != nil {
			return err
		}
		if end, err = time.Parse(servemetf.TimeFormat, (*args.Serveme).EndsAt); err != nil {
			return err
		}

		randBytes := make([]byte, 6)
		rand.Read(randBytes)
		*args.RconPwd = base64.URLEncoding.EncodeToString(randBytes)

		reservation = servemetf.Reservation{
			StartsAt:    start.Format(servemetf.TimeFormat),
			EndsAt:      end.Format(servemetf.TimeFormat),
			ServerID:    (*args.Serveme).Server.ID,
			WhitelistID: 1,
			RCON:        *args.RconPwd,
			Password:    "foobar",
		}

		context = helpers.GetServemeContextIP(chelpers.GetIPAddr(so.Request))
		resp, err := context.Create(reservation, so.Token.Claims.(*chelpers.TF2StadiumClaims).SteamID)
		if err != nil || resp.Reservation.Errors != nil {
			if err != nil {
				logrus.Error(err)
			} else {
				logrus.Error(resp.Reservation.Errors)
			}

			return errors.New("Couldn't get serveme reservation")
		}

		*args.Server = resp.Reservation.Server.IPAndPort
		reservation = resp.Reservation
	} else if *args.ServerType == "storedServer" {
		if *args.Server == "" {
			return errors.New("No server ID given")
		}

		id, err := strconv.ParseUint(*args.Server, 10, 64)
		if err != nil {
			return err
		}
		server, err := gameserver.GetStoredServer(uint(id))
		if err != nil {
			return err
		}
		*args.Server = server.Address
		*args.RconPwd = server.RCONPassword
	} else { // *args.ServerType == "server"
		if args.RconPwd == nil || *args.RconPwd == "" {
			return errors.New("RCON Password cannot be empty")
		}
		if args.Server == nil || *args.Server == "" {
			return errors.New("Server Address cannot be empty")
		}
	}

	var count int

	lobbyType := playermap[*args.Type]
	db.DB.Model(&gameserver.ServerRecord{}).Where("host = ?", *args.Server).Count(&count)
	if count != 0 {
		return errors.New("A lobby is already using this server.")
	}

	randBytes := make([]byte, 6)
	rand.Read(randBytes)
	serverPwd := base64.URLEncoding.EncodeToString(randBytes)

	//TODO what if playermap[lobbytype] is nil?
	info := gameserver.ServerRecord{
		Host:           *args.Server,
		RconPassword:   *args.RconPwd,
		ServerPassword: serverPwd,
	}

	lob := lobby.NewLobby(*args.Map, lobbyType, *args.League, info, *args.WhitelistID, *args.Mumble, steamGroup)

	if args.TwitchWhitelistSubscribers || args.TwitchWhitelistFollowers {
		if p.TwitchName == "" {
			return errors.New("Please connect your twitch account first.")
		}

		lob.TwitchChannel = p.TwitchName
		if args.TwitchWhitelistFollowers {
			lob.TwitchRestriction = lobby.TwitchFollowers
		} else {
			lob.TwitchRestriction = lobby.TwitchSubscribers
		}
	}

	lob.Discord = args.Discord != nil
	if lob.Discord {
		if !reDiscordInvite.MatchString(*args.Discord.RedChannel) || !reDiscordInvite.MatchString(*args.Discord.BluChannel) {
			return errors.New("Invalid Discord invite URL")
		}

		lob.DiscordRedChannel = *args.Discord.RedChannel
		lob.DiscordBluChannel = *args.Discord.BluChannel
	}

	lob.RegionLock = args.RegionLock
	lob.CreatedBySteamID = p.SteamID
	lob.RegionCode, lob.RegionName = helpers.GetRegion(*args.Server)
	if (lob.RegionCode == "" || lob.RegionName == "") && config.Constants.GeoIP {
		if reservation.ID != 0 {
			err := context.Delete(reservation.ID, p.SteamID)
			for err != nil {
				err = context.Delete(reservation.ID, p.SteamID)
			}
		} else if *args.ServerType == "storedServer" {
			gameserver.PutStoredServer(*args.Server)
		}

		return errors.New("Couldn't find the region for this server.")
	}

	if lobby.MapRegionFormatExists(lob.MapName, lob.RegionCode, lob.Type) {
		if reservation.ID != 0 {
			err := context.Delete(reservation.ID, p.SteamID)
			for err != nil {
				err = context.Delete(reservation.ID, p.SteamID)
			}
		} else if *args.ServerType == "storedServer" {
			gameserver.PutStoredServer(*args.Server)
		}

		return errors.New("Your region already has a lobby with this map and format.")
	}

	if *args.ServerType == "serveme" {
		lob.ServemeID = reservation.ID
	}

	lob.Save()
	lob.CreateLock()

	if *args.ServerType == "serveme" {
		now := time.Now()

		for {
			status, err := context.Status(reservation.ID, p.SteamID)
			if err != nil {
				logrus.Error(err)
			}
			if status == "Ready" {
				break
			}

			time.Sleep(10 * time.Second)
			if time.Since(now) >= 5*time.Minute {
				lob.Delete()
				return errors.New("Couldn't get Serveme reservation, try another server.")
			}
		}

		lob.ServemeCheck(context)
	}

	err := lob.SetupServer()
	if err != nil { //lobby setup failed, delete lobby and corresponding server record
		lob.Delete()
		return err
	}

	lob.SetState(lobby.Waiting)

	if args.Requirements != nil {
		for class, requirement := range (*args.Requirements).Classes {
			if requirement.Restricted.Blu {
				newRequirement("blu", class, requirement, lob)
			}
			if requirement.Restricted.Red {
				newRequirement("red", class, requirement, lob)
			}
		}
		if args.Requirements.General.Hours != 0 || args.Requirements.General.Lobbies != 0 {
			for i := 0; i < 2*format.NumberOfClassesMap[lob.Type]; i++ {
				req := &lobby.Requirement{
					LobbyID: lob.ID,
					Hours:   args.Requirements.General.Hours,
					Lobbies: args.Requirements.General.Lobbies,
					Slot:    i,
				}
				req.Save()
			}

		}
	}

	if *args.Password != "" {
		for i := 0; i < 2*format.NumberOfClassesMap[lob.Type]; i++ {
			req := &lobby.Requirement{
				LobbyID:  lob.ID,
				Slot:     i,
				Password: *args.Password,
			}
			req.Save()
		}
	}

	chat.NewBotMessage(fmt.Sprintf("Lobby created by %s", p.Alias()), int(lob.ID)).Send()

	lobby.BroadcastLobbyList()
	return newResponse(
		struct {
			ID uint `json:"id"`
		}{lob.ID})
}

func (Lobby) LobbyServerReset(so *wsevent.Client, args struct {
	ID *uint `json:"id"`
}) interface{} {

	player := chelpers.GetPlayer(so.Token)
	lob, tperr := lobby.GetLobbyByID(*args.ID)

	if player.SteamID != lob.CreatedBySteamID && (player.Role != helpers.RoleAdmin && player.Role != helpers.RoleMod) {
		return errors.New("You are not authorized to reset server.")
	}

	if tperr != nil {
		return tperr
	}

	if lob.State == lobby.Ended {
		return errors.New("Lobby has ended")
	}

	if err := rpc.ReExecConfig(lob.ID, false); err != nil {
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
	db.DB.Model(&gameserver.ServerRecord{}).Where("host = ?", *args.Server).Count(&count)
	if count != 0 {
		return errors.New("A lobby is already using this server.")
	}

	info := &gameserver.ServerRecord{
		Host:         *args.Server,
		RconPassword: *args.Rconpwd,
	}
	db.DB.Save(info)
	defer db.DB.Delete(info)

	err := rpc.VerifyInfo(*info)
	if err != nil {
		return err
	}

	return emptySuccess
}

func (Lobby) LobbyClose(so *wsevent.Client, args struct {
	Id *uint `json:"id"`
}) interface{} {

	player := chelpers.GetPlayer(so.Token)
	lob, tperr := lobby.GetLobbyByIDServer(uint(*args.Id))
	if tperr != nil {
		return tperr
	}

	if player.SteamID != lob.CreatedBySteamID && !(player.Role == helpers.RoleAdmin || player.Role == helpers.RoleMod) {
		return errors.New("Player not authorized to close lobby.")

	}

	if lob.State == lobby.Ended {
		return errors.New("Lobby already closed.")
	}

	lob.Close(true, false)

	notify := fmt.Sprintf("Lobby closed by %s", player.Alias())
	chat.SendNotification(notify, int(lob.ID))

	return emptySuccess
}

// for rate limiting notifs (like lobby-almost-ready discord
// notifs). Bad because it gets restart on Helen restart... but easy
// for now
var lobbyJoinLastNotif = make(map[uint]time.Time)
var notifThreshold = map[format.Format]int{
	format.Highlander: 13,
	format.Prolander:  8,
	format.Sixes:      8,
	format.Fours:      5,
	format.Ultiduo:    2,
	format.Bball:      2,
	format.Debug:      1,
}

func (Lobby) LobbyJoin(so *wsevent.Client, args struct {
	Id       *uint   `json:"id"`
	Class    *string `json:"class"`
	Team     *string `json:"team" valid:"red,blu"`
	Password *string `json:"password" empty:"-"`
}) interface{} {

	p := chelpers.GetPlayer(so.Token)
	if banned, until := p.IsBannedWithTime(player.BanJoin); banned {
		ban, _ := p.GetActiveBan(player.BanJoin)
		return fmt.Errorf("You have been banned from joining lobbies till %s (%s)", until.Format(time.RFC822), ban.Reason)
	}

	//logrus.Debug("id %d class %s team %s", *args.Id, *args.Class, *args.Team)
	lob, tperr := lobby.GetLobbyByID(*args.Id)

	if tperr != nil {
		return tperr
	}

	if lob.Mumble {
		if banned, until := p.IsBannedWithTime(player.BanJoinMumble); banned {
			ban, _ := p.GetActiveBan(player.BanJoinMumble)
			return fmt.Errorf("You have been banned from joining Mumble lobbies till %s (%s)", until.Format(time.RFC822), ban.Reason)
		}
	}

	if lob.State == lobby.Ended {
		return errors.New("Cannot join a closed lobby.")

	}
	if lob.State == lobby.Initializing {
		return errors.New("Lobby is being setup right now.")
	}

	if lob.RegionLock {
		region, _ := helpers.GetRegion(chelpers.GetIPAddr(so.Request))
		if region != lob.RegionCode {
			return errors.New("This lobby is region locked.")
		}
	}

	//Check if player is in the same lobby
	var sameLobby bool
	if id, err := p.GetLobbyID(false); err == nil && id == *args.Id {
		sameLobby = true
	}

	slot, tperr := format.GetSlot(lob.Type, *args.Team, *args.Class)
	if tperr != nil {
		return tperr
	}

	if prevId, _ := p.GetLobbyID(false); prevId != 0 && !sameLobby {
		lob, _ := lobby.GetLobbyByID(prevId)
		hooks.AfterLobbyLeave(lob, p, false, false)
	}

	tperr = lob.AddPlayer(p, slot, *args.Password)

	if tperr != nil {
		return tperr
	}

	if !sameLobby {
		hooks.AfterLobbyJoin(so, lob, p)
	}

	playersCnt := lob.GetPlayerNumber()
	lastNotif, timerExists := lobbyJoinLastNotif[lob.ID]
	if playersCnt >= notifThreshold[lob.Type] && !lob.IsEnoughPlayers(playersCnt) && (!timerExists || time.Since(lastNotif).Minutes() > 5) {
		lob.DiscordNotif(fmt.Sprintf("Almost ready [%d/%d]", playersCnt, lob.RequiredPlayers()))
		lobbyJoinLastNotif[lob.ID] = time.Now()
	}

	//check if lobby isn't already in progress (which happens when the player is subbing)
	lob.Lock()
	if lob.IsEnoughPlayers(playersCnt) && lob.State != lobby.InProgress && lob.State != lobby.ReadyingUp {
		lob.State = lobby.ReadyingUp
		lob.ReadyUpTimestamp = time.Now().Unix() + 30
		lob.Save()

		helpers.GlobalWait.Add(1)
		time.AfterFunc(time.Second*30, func() {
			state := lob.CurrentState()
			//if all player's haven't readied up,
			//remove unreadied players and unready the
			//rest.
			//don't do this when:
			//  lobby.State == Waiting (someone already unreadied up, so all players have been unreadied)
			// lobby.State == InProgress (all players have readied up, so the lobby has started)
			// lobby.State == Ended (the lobby has been closed)
			if state != lobby.Waiting && state != lobby.InProgress && state != lobby.Ended {
				lob.SetState(lobby.Waiting)
				removeUnreadyPlayers(lob)
				lob.UnreadyAllPlayers()
				//get updated lobby object
				lob, _ = lobby.GetLobbyByID(lob.ID)
				lobby.BroadcastLobby(lob)
			}
			helpers.GlobalWait.Done()
		})

		room := fmt.Sprintf("%s_private",
			hooks.GetLobbyRoom(lob.ID))
		broadcaster.SendMessageToRoom(room, "lobbyReadyUp",
			struct {
				Timeout int `json:"timeout"`
			}{30})
		lobby.BroadcastLobbyList()
	}
	lob.Unlock()

	if lob.State == lobby.InProgress { //this happens when the player is a substitute
		db.DB.Preload("ServerInfo").First(lob, lob.ID)
		so.EmitJSON(helpers.NewRequest("lobbyStart", lobby.DecorateLobbyConnect(lob, p, slot)))
	}

	return emptySuccess
}

//get list of unready players, remove them from lobby (and add them as spectators)
//plus, call the after lobby leave hook for each player removed
func removeUnreadyPlayers(lobby *lobby.Lobby) {
	players := lobby.GetUnreadyPlayers()
	lobby.RemoveUnreadyPlayers(true)

	for _, player := range players {
		hooks.AfterLobbyLeave(lobby, player, false, true)
	}
}

func (Lobby) LobbySpectatorJoin(so *wsevent.Client, args struct {
	Id *uint `json:"id"`
}) interface{} {

	lob, err := lobby.GetLobbyByID(*args.Id)

	if err != nil {
		return err
	}

	player := chelpers.GetPlayer(so.Token)
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

	hooks.AfterLobbySpec(socket.AuthServer, so, player, lob)
	lobby.BroadcastLobbyToUser(lob, player.SteamID)
	return emptySuccess
}

func removePlayerFromLobby(lobbyId uint, steamId string) (*lobby.Lobby, *player.Player, error) {
	player, err := player.GetPlayerBySteamID(steamId)
	if err != nil {
		return nil, nil, err
	}

	lob, err := lobby.GetLobbyByID(lobbyId)
	if err != nil {
		return nil, nil, err
	}

	switch lob.State {
	case lobby.InProgress:
		lob.Substitute(player)
		return lob, player, lob.AddSpectator(player)
	case lobby.Ended:
		return lob, player, errors.New("Lobby has closed.")
	}

	_, err = lob.GetPlayerSlot(player)
	if err != nil {
		return lob, player, errors.New("Player not playing")
	}

	if err := lob.RemovePlayer(player); err != nil {
		return lob, player, err
	}

	return lob, player, lob.AddSpectator(player)
}

func playerCanKick(lobbyId uint, steamId string) (bool, error) {
	lob, err := lobby.GetLobbyByID(lobbyId)
	if err != nil {
		return false, err
	}

	player, err := player.GetPlayerBySteamID(steamId)
	if err != nil {
		return false, err
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
	selfSteamId := so.Token.Claims.(*chelpers.TF2StadiumClaims).SteamID

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

	hooks.AfterLobbyLeave(lob, player, true, false)

	// broadcaster.SendMessage(steamId, "sendNotification",
	// 	fmt.Sprintf(`{"notification": "You have been removed from Lobby #%d"}`, *args.Id))

	return emptySuccess
}

func (Lobby) LobbyBan(so *wsevent.Client, args struct {
	Id      *uint   `json:"id"`
	Steamid *string `json:"steamid"`
}) interface{} {

	steamId := *args.Steamid
	selfSteamId := so.Token.Claims.(*chelpers.TF2StadiumClaims).SteamID

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

	hooks.AfterLobbyLeave(lob, player, true, false)

	// broadcaster.SendMessage(steamId, "sendNotification",
	// 	fmt.Sprintf(`{"notification": "You have been removed from Lobby #%d"}`, *args.Id))

	return emptySuccess
}

func (Lobby) LobbyLeave(so *wsevent.Client, args struct {
	Id *uint `json:"id"`
}) interface{} {

	steamId := so.Token.Claims.(*chelpers.TF2StadiumClaims).SteamID

	lob, player, tperr := removePlayerFromLobby(*args.Id, steamId)
	if tperr != nil {
		return tperr
	}

	hooks.AfterLobbyLeave(lob, player, false, false)

	return emptySuccess
}

func (Lobby) LobbySpectatorLeave(so *wsevent.Client, args struct {
	Id *uint `json:"id"`
}) interface{} {

	player := chelpers.GetPlayer(so.Token)
	lob, tperr := lobby.GetLobbyByID(*args.Id)
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
	so.EmitJSON(helpers.NewRequest("lobbyListData", lobby.DecorateLobbyListData(lobby.GetWaitingLobbies(), false)))

	return emptySuccess
}

func (Lobby) LobbyChangeOwner(so *wsevent.Client, args struct {
	ID      *uint   `json:"id"`
	SteamID *string `json:"steamid"`
}) interface{} {
	lob, err := lobby.GetLobbyByID(*args.ID)
	if err != nil {
		return err
	}

	// current owner
	player1 := chelpers.GetPlayer(so.Token)
	if lob.CreatedBySteamID != player1.SteamID {
		return errors.New("You aren't authorized to change lobby owner.")
	}

	// to be owner
	player2, err := player.GetPlayerBySteamID(*args.SteamID)
	if err != nil {
		return err
	}

	lob.CreatedBySteamID = player2.SteamID
	lob.Save()
	lobby.BroadcastLobby(lob)
	lobby.BroadcastLobbyList()
	chat.NewBotMessage(fmt.Sprintf("Lobby leader changed to %s", player2.Alias()), int(*args.ID)).Send()

	return emptySuccess
}

func (Lobby) LobbySetRequirement(so *wsevent.Client, args struct {
	ID *uint `json:"id"` // lobby ID

	Slot     *int         `json:"slot"` // -1 if to set for all slots
	Type     *string      `json:"type"`
	Value    *json.Number `json:"value"`
	Password *string      `json:"password" empty:"-"`
}) interface{} {

	lob, err := lobby.GetLobbyByID(*args.ID)
	if err != nil {
		return err
	}

	player := chelpers.GetPlayer(so.Token)
	if lob.CreatedBySteamID != player.SteamID {
		return errors.New("Only lobby owners can change requirements.")
	}

	if !(*args.Slot >= 0 && *args.Slot < 2*format.NumberOfClassesMap[lob.Type]) {
		return errors.New("Invalid slot.")
	}

	req, err := lob.GetSlotRequirement(*args.Slot)
	if err != nil { //requirement doesn't exist. create one
		req.Slot = *args.Slot
		req.LobbyID = *args.ID
	}

	var n int64
	var f float64
	var deleted = false

	switch *args.Type {
	case "hours":
		n, err = args.Value.Int64()
		if n == 0 {
			// hours being 0 is a special case to allow players past the
			// default 150 hr limit; so have to actually delete in this case
			db.DB.Delete(&req)
			deleted = true
		} else {
			req.Hours = int(n)
		}
	case "lobbies":
		n, err = args.Value.Int64()
		req.Lobbies = int(n)
	case "reliability":
		f, err = args.Value.Float64()
		req.Reliability = f
	case "password":
		req.Password = *args.Password
	default:
		return errors.New("Invalid requirement type.")
	}

	if err != nil {
		return errors.New("Invalid requirement.")
	}

	if !deleted {
		req.Save()
	}
	lobby.BroadcastLobby(lob)
	lobby.BroadcastLobbyList()

	return emptySuccess
}

func (Lobby) LobbySetTeamName(so *wsevent.Client, args struct {
	Id      uint   `json:"id"`
	Team    string `json:"team"`
	NewName string `json:"name"`
}) interface{} {
	player := chelpers.GetPlayer(so.Token)

	lob, err := lobby.GetLobbyByID(args.Id)
	if err != nil {
		return err
	}

	if player.SteamID != lob.CreatedBySteamID && (player.Role != helpers.RoleAdmin && player.Role != helpers.RoleMod) {
		return errors.New("You aren't authorized to do this.")
	}

	if len(args.NewName) == 0 || len(args.NewName) > 12 {
		return errors.New("team name must be between 1-12 characters long.")
	}

	if args.Team == "red" {
		lob.RedTeamName = args.NewName
	} else if args.Team == "blu" {
		lob.BluTeamName = args.NewName
	} else {
		return errors.New("team must be red or blu.")
	}

	lob.Save()
	lobby.BroadcastLobby(lob)
	return emptySuccess
}

func (Lobby) LobbyRemoveTwitchRestriction(so *wsevent.Client, args struct {
	ID uint `json:"id"`
}) interface{} {
	player := chelpers.GetPlayer(so.Token)

	lob, err := lobby.GetLobbyByID(args.ID)
	if err != nil {
		return err
	}

	if player.SteamID != lob.CreatedBySteamID && (player.Role != helpers.RoleAdmin && player.Role != helpers.RoleMod) {
		return errors.New("You aren't authorized to do this.")
	}

	lob.TwitchChannel = ""
	lob.Save()

	lobby.BroadcastLobby(lob)
	lobby.BroadcastLobbyList()

	return emptySuccess
}

func (Lobby) LobbyRemoveSteamRestriction(so *wsevent.Client, args struct {
	ID uint `json:"id"`
}) interface{} {
	player := chelpers.GetPlayer(so.Token)

	lob, err := lobby.GetLobbyByID(args.ID)
	if err != nil {
		return err
	}

	if player.SteamID != lob.CreatedBySteamID && (player.Role != helpers.RoleAdmin && player.Role != helpers.RoleMod) {
		return errors.New("You aren't authorized to do this.")
	}

	lob.PlayerWhitelist = ""
	lob.Save()

	lobby.BroadcastLobby(lob)
	lobby.BroadcastLobbyList()

	return emptySuccess
}

func (Lobby) LobbyRemoveRegionLock(so *wsevent.Client, args struct {
	ID uint `json:"id"`
}) interface{} {
	player := chelpers.GetPlayer(so.Token)

	lob, err := lobby.GetLobbyByID(args.ID)
	if err != nil {
		return err
	}

	if player.SteamID != lob.CreatedBySteamID && (player.Role != helpers.RoleAdmin && player.Role != helpers.RoleMod) {
		return errors.New("You aren't authorized to do this.")
	}

	lob.RegionLock = false
	lob.Save()

	lobby.BroadcastLobby(lob)
	lobby.BroadcastLobbyList()

	return emptySuccess

}

func (Lobby) LobbyShuffle(so *wsevent.Client, args struct {
	Id uint `json:"id"`
}) interface{} {
	player := chelpers.GetPlayer(so.Token)

	lob, err := lobby.GetLobbyByID(args.Id)
	if err != nil {
		return err
	}

	if player.SteamID != lob.CreatedBySteamID && (player.Role != helpers.RoleAdmin && player.Role != helpers.RoleMod) {
		return errors.New("You aren't authorized to shuffle this lobby.")
	}

	if err = lob.ShuffleAllSlots(); err != nil {
		return err
	}

	room := fmt.Sprintf("%s_private", hooks.GetLobbyRoom(args.Id))
	broadcaster.SendMessageToRoom(room, "lobbyShuffled", args)
	chat.NewBotMessage(fmt.Sprintf("Lobby shuffled by %s", player.Alias()), int(args.Id)).Send()
	return emptySuccess
}

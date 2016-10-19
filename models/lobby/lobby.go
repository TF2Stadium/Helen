// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package lobby

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/chat"
	"github.com/TF2Stadium/Helen/models/gameserver"
	"github.com/TF2Stadium/Helen/models/lobby/format"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/Helen/models/rpc"
	"github.com/TF2Stadium/PlayerStatsScraper/steamid"
	"github.com/TF2Stadium/logstf"
	"github.com/TF2Stadium/servemetf"
	"github.com/jinzhu/gorm"
)

type State int

const (
	Initializing State = 0
	Waiting      State = 1
	ReadyingUp   State = 2
	InProgress   State = 3
	Ended        State = 5
)

var (
	ErrLobbyNotFound   = errors.New("Could not find lobby with given ID")
	ErrLobbyBan        = errors.New("You have been banned from this lobby")
	ErrBadSlot         = errors.New("That slot does not exist")
	ErrFilled          = errors.New("That slot has been filled")
	ErrNotWhitelisted  = errors.New("You are not allowed in this lobby")
	ErrInvalidPassword = errors.New("Incorrect slot password")
	ErrNeedsSub        = errors.New("This slot needs a substitute")

	ErrReqHours       = errors.New("You do not have sufficient hours to join that slot")
	ErrReqLobbies     = errors.New("You have not played sufficient lobbies to join that slot")
	ErrReqReliability = errors.New("You have insufficient reliability to join that slot")
)

// Represents an occupied player slot in a lobby
type LobbySlot struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	LobbyID  uint //ID of the player occupying the slot
	PlayerID uint //Slot number
	Slot     int  //Denotes if the player is ready
	Ready    bool //Denotes if the player is in game
	InGame   bool //true if the player is in the game server
	InMumble bool //true if the player is in the mumble channel for the lobby
	NeedsSub bool //true if the slot needs a subtitute player
}

//DeleteUnusedServerRecords checks all server records in the DB and deletes them if
//the corresponsing lobby is closed
func DeleteUnusedServers() {
	serverInfoIDs := []uint{}
	db.DB.Model(&gameserver.ServerRecord{}).Pluck("id", &serverInfoIDs)
	for _, id := range serverInfoIDs {
		lobby := &Lobby{}
		err := db.DB.Where("server_info_id = ?", id).First(lobby).Error

		if err != nil || lobby.State == Ended {
			db.DB.Model(&gameserver.ServerRecord{}).Where("id = ?", id).Delete(&gameserver.ServerRecord{})
		}
	}
}

type TwitchRestriction int

const (
	TwitchSubscribers TwitchRestriction = iota
	TwitchFollowers
)

func (t TwitchRestriction) String() string {
	if t == TwitchSubscribers {
		return "subscribers"
	}

	return "followers"
}

//Lobby represents a Lobby
type Lobby struct {
	gorm.Model
	State State

	Mode    string        // Game Mode
	MapName string        // Map Name
	Type    format.Format // League config used
	League  string

	RegionCode string // Region Code ("na", "eu", etc)
	RegionName string // Region Name ("North America", "Europe", etc)

	Mumble bool // Whether mumble is required

	Slots []LobbySlot `gorm:"ForeignKey:LobbyID"` // List of occupied slots

	RegionLock        bool
	PlayerWhitelist   string            // URL of steam group
	TwitchChannel     string            // twitch channel, slots will be restricted
	TwitchRestriction TwitchRestriction // restricted to either followers or subs
	ServemeID         int               // if serveme was used to get this server, stores the server ID

	// TF2 Server Info
	ServerInfo   gameserver.ServerRecord `gorm:"ForeignKey:ServerInfoID"`
	ServerInfoID uint

	Whitelist string //whitelist.tf ID

	Spectators    []player.Player `gorm:"many2many:spectators_players_lobbies"` // List of spectators
	BannedPlayers []player.Player `gorm:"many2many:banned_players_lobbies"`     // List of Banned Players

	CreatedBySteamID string // SteamID of the lobby leader/creator

	ReadyUpTimestamp int64 // (Unix) Timestamp at which the ready up timeout started
	MatchEnded       bool  // if true, the lobby ended with the match ending in the game server
	LogstfID         int   // logs.tf id (only when match ends)
}

func getGamemode(mapName string, lobbyType format.Format) string {
	switch {
	case strings.HasPrefix(mapName, "koth"):
		if lobbyType == format.Ultiduo {
			return "ultiduo"
		}

		return "koth"

	case strings.HasPrefix(mapName, "ctf"):
		if lobbyType == format.Bball {
			return "bball"
		}

		return "ctf"

	case strings.HasPrefix(mapName, "cp"):
		if mapName == "cp_gravelpit" {
			return "a/d"
		}

		return "5cp"

	case strings.HasPrefix(mapName, "pl"):
		return "payload"

	case strings.HasPrefix(mapName, "arena"):
		return "arena"
	}

	return "unknown"
}

func MapRegionFormatExists(mapName, region string, lobbytype format.Format) bool {
	var count int

	db.DB.Model(&Lobby{}).Where("map_name = ? AND region_code = ? AND type = ? AND state = ?", mapName, region, lobbytype, Waiting).Count(&count)
	return count != 0
}

// Returns a new lobby object with the given parameters
// Call CreateLock after saving this lobby.
func NewLobby(mapName string, lobbyType format.Format, league string, serverInfo gameserver.ServerRecord, whitelist string, mumble bool, whitelistGroup string) *Lobby {
	lobby := &Lobby{
		Mode:            getGamemode(mapName, lobbyType),
		Type:            lobbyType,
		State:           Initializing,
		League:          league,
		MapName:         mapName,
		Whitelist:       whitelist, // that's a strange line
		Mumble:          mumble,
		ServerInfo:      serverInfo,
		PlayerWhitelist: whitelistGroup,
	}

	// Must specify CreatedBy manually if the lobby is created by a player
	return lobby
}

//Delete removes the lobby object from the database.
//Closed lobbies aren't deleted, this function is used for
//lobbies where the game server had an error while being setup.
func (lobby *Lobby) Delete() {
	var count int

	db.DB.Model(&gameserver.ServerRecord{}).Where("host = ?", lobby.ServerInfo.Host).Count(&count)
	if count != 0 {
		gameserver.PutStoredServer(lobby.ServerInfo.Host)
	}

	if lobby.ServemeID != 0 {
		context := helpers.GetServemeContext(lobby.ServerInfo.Host)
		err := context.Delete(lobby.ServemeID, lobby.CreatedBySteamID)
		for err != nil {
			err = context.Delete(lobby.ServemeID, lobby.CreatedBySteamID)
		}
	}

	db.DB.Delete(lobby)
	db.DB.Delete(&lobby.ServerInfo)

	lobby.deleteLock()
}

//GetWaitingLobbies returns a list of lobby objects that haven't been filled yet
func GetWaitingLobbies() (lobbies []*Lobby) {
	db.DB.Where("state = ?", Waiting).Order("id desc").Find(&lobbies)
	return
}

//CurrentState returns the lobby's current state.
//It's meant to be used for old lobby objects which might have their state change while the
//object hasn't been updated.
func (l *Lobby) CurrentState() State {
	var state int
	db.DB.DB().QueryRow("SELECT state FROM lobbies WHERE id = $1", l.ID).Scan(&state)
	return State(state)
}

func (l *Lobby) SetState(s State) {
	db.DB.Model(&Lobby{}).Where("id = ?", l.ID).UpdateColumn("state", s)
	l.State = s
}

//ServemeCheck checks the status of the serveme reservation for the lobby
//(if any) every 10 seconds in a goroutine, and closes the lobby if it has ended
func (l *Lobby) ServemeCheck(context *servemetf.Context) {
	go func() {
		for {
			ended, err := context.Ended(l.ServemeID, l.CreatedBySteamID)
			if err != nil {
				logrus.Error(err)
			}
			if ended {
				if l.CurrentState() != Ended {
					chat.SendNotification("Lobby Closed (Serveme reservation ended.)", int(l.ID))
					l.Close(true, false)
				}
				return
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

func RestoreServemeChecks() {
	var ids []uint
	db.DB.Model(&Lobby{}).Where("state <> ? AND serveme_id <> 0", Ended).Pluck("id", &ids)

	for _, id := range ids {
		lobby, _ := GetLobbyByIDServer(id)
		context := helpers.GetServemeContext(lobby.ServerInfo.Host)
		lobby.ServemeCheck(context)
	}
}

//GetPlayerSlotObj returns the LobbySlot object if the given player occupies a slot in the lobby.
func (lobby *Lobby) GetPlayerSlotObj(player *player.Player) (*LobbySlot, error) {
	slotObj := &LobbySlot{}

	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).First(slotObj).Error

	return slotObj, err
}

//GetPlayerSlot returns the slot number if the player occupies a slot int eh lobby
func (lobby *Lobby) GetPlayerSlot(player *player.Player) (int, error) {
	slotObj, err := lobby.GetPlayerSlotObj(player)

	return slotObj.Slot, err
}

//GetPlayerIDBySlot returns the ID of the player occupying the slot number
func (lobby *Lobby) GetPlayerIDBySlot(slot int) (uint, error) {
	slotObj := &LobbySlot{}

	err := db.DB.Where("lobby_id = ? AND slot = ?", lobby.ID, slot).First(slotObj).Error

	return uint(slotObj.PlayerID), err
}

//Save saves changes made to lobby object to the DB
func (lobby *Lobby) Save() error {
	var err error
	if db.DB.NewRecord(lobby) {
		err = db.DB.Create(lobby).Error
	} else {
		err = db.DB.Save(lobby).Error
	}

	lobby.OnChange(true)
	return err
}

//GetLobbyByIdServer returns the lobby object, plus the ServerInfo object inside it
func GetLobbyByIDServer(id uint) (*Lobby, error) {
	lob := &Lobby{}
	err := db.DB.Preload("ServerInfo").First(lob, id).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrLobbyNotFound
	}

	return lob, nil
}

//GetLobbyByID returns lobby object, without the ServerInfo object inside it.
func GetLobbyByID(id uint) (*Lobby, error) {
	lob := &Lobby{}
	err := db.DB.First(lob, id).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrLobbyNotFound
	}

	return lob, err
}

//HasPlayer returns true if the given player occupies a slot in the lobby
func (lobby *Lobby) HasPlayer(player *player.Player) bool {
	var count int
	db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).Count(&count)

	return count != 0
}

//SlotNeedsSubstitute returns true if the given slot needs a substitute
func (lobby *Lobby) SlotNeedsSubstitute(slot int) (needsSub bool) {
	//use database/sql API, since it's simpler here
	db.DB.DB().QueryRow("SELECT needs_sub FROM lobby_slots WHERE lobby_id = $1 AND slot = $2", lobby.ID, slot).Scan(&needsSub)
	return
}

//FillSubstitute marks the substitute reocrd for the given slot as true, and Broadcasts the updated sub list
func (lobby *Lobby) FillSubstitute(slot int) error {
	err := db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND slot = ?", lobby.ID, slot).UpdateColumn("needs_sub", false).Error
	BroadcastSubList()
	return err
}

//IsPlayerBanned returns whether the given player is banned from joining the lobby.
func (lobby *Lobby) IsPlayerBanned(player *player.Player) bool {
	var count int
	db.DB.Table("banned_players_lobbies").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).Count(&count)
	return count != 0
}

//IsSlotOccupied returns whether the given slot is occupied by a player.
func (lobby *Lobby) IsSlotOccupied(slot int) bool {
	var count int
	db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND slot = ?", lobby.ID, slot).Count(&count)
	return count != 0
}

//AddPlayer adds the given player to lobby, If the player occupies a slot in the lobby already, switch slots.
//If the player is in another lobby, removes them from that lobby before adding them.
func (lobby *Lobby) AddPlayer(p *player.Player, slot int, password string) error {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */

	//Check if player is banned
	if lobby.IsPlayerBanned(p) {
		return ErrLobbyBan
	}

	if slot >= 2*format.NumberOfClassesMap[lobby.Type] || slot < 0 {
		return ErrBadSlot
	}

	isSubstitution := lobby.SlotNeedsSubstitute(slot)

	//Check whether the slot is occupied
	if !isSubstitution && lobby.IsSlotOccupied(slot) {
		return ErrFilled
	}

	if lobby.HasSlotRequirement(slot) {
		//check if player fits the requirements for the slot
		if ok, err := lobby.FitsRequirements(p, slot); !ok {
			return err
		}

		req, _ := lobby.GetSlotRequirement(slot)
		if password != req.Password {
			return ErrInvalidPassword
		}
	}

	var slotChange bool
	//Check if the player is currently in another lobby
	if currLobbyID, err := p.GetLobbyID(false); err == nil {
		if currLobbyID != lobby.ID {
			//if the player is in a different lobby, remove them from that lobby
			//plus substitute them
			curLobby, _ := GetLobbyByID(currLobbyID)

			if curLobby.State == InProgress {
				curLobby.Substitute(p)
			} else {
				curLobby.Lock()
				db.DB.Where("player_id = ? AND lobby_id = ?", p.ID, curLobby.ID).Delete(&LobbySlot{})
				curLobby.Unlock()

			}

		} else { //player is in the same lobby, they're changing their slots
			//assign the player to a new slot
			if isSubstitution {
				//the slot needs a substitute (which happens when the lobby is in progress),
				//so players already in the lobby cannot fill it.
				return ErrNeedsSub
			}
			lobby.Lock()
			db.DB.Where("player_id = ? AND lobby_id = ?", p.ID, lobby.ID).Delete(&LobbySlot{})
			lobby.Unlock()

			slotChange = true
		}
	}

	if !slotChange {
		//check if the player is in the steam group whitelist
		url := fmt.Sprintf(`http://steamcommunity.com/groups/%s/memberslistxml/?xml=1`,
			lobby.PlayerWhitelist)

		if lobby.PlayerWhitelist != "" && !helpers.IsWhitelisted(p.SteamID, url) {
			return ErrNotWhitelisted
		}

		//check if player has been subbed to the twitch channel (if any)
		//allow channel owners
		if lobby.TwitchChannel != "" && p.TwitchName != lobby.TwitchChannel {
			//check if player has connected their twitch account
			if p.TwitchAccessToken == "" {
				return errors.New("You need to connect your Twitch Account first to join the lobby.")
			}
			if lobby.TwitchRestriction == TwitchSubscribers && !p.IsSubscribed(lobby.TwitchChannel) {
				return fmt.Errorf("You aren't subscribed to %s", lobby.TwitchChannel)
			}
			if lobby.TwitchRestriction == TwitchFollowers && !p.IsFollowing(lobby.TwitchChannel) {
				return fmt.Errorf("You aren't following %s", lobby.TwitchChannel)
			}
		}
	}

	// Check if player is a substitute (the slot needs a subtitute)
	if isSubstitution {
		//get previous slot, to kick them from game
		prevPlayerID, _ := lobby.GetPlayerIDBySlot(slot)
		prevPlayer, _ := player.GetPlayerByID(prevPlayerID)

		lobby.Lock()
		db.DB.Where("lobby_id = ? AND slot = ?", lobby.ID, slot).Delete(&LobbySlot{})
		lobby.Unlock()

		go func() {
			//kicks previous slot occupant if they're in-game, resets their !rep count, removes them from the lobby
			rpc.DisallowPlayer(lobby.ID, prevPlayer.SteamID, prevPlayer.ID)
			BroadcastSubList() //since the sub slot has been deleted, broadcast the updated substitute list
			//notify players in game server of subtitute
			class, team, _ := format.GetSlotTeamClass(lobby.Type, slot)
			rpc.Say(lobby.ID, fmt.Sprintf("Substitute found for %s %s: %s (%s)", team, class, p.Name, p.SteamID))
		}()
		//allow player in mumble
	}

	//try to remove them from spectators
	lobby.RemoveSpectator(p, true)

	newSlotObj := &LobbySlot{
		PlayerID: p.ID,
		LobbyID:  lobby.ID,
		Slot:     slot,
	}

	lobby.Lock()
	db.DB.Create(newSlotObj)
	lobby.Unlock()
	if !slotChange {
		if p.TwitchName != "" {
			rpc.TwitchBotAnnouce(p.TwitchName, lobby.ID)
		}
	}

	lobby.OnChange(true)
	p.SetMumbleUsername(lobby.Type, slot)

	return nil
}

//RemovePlayer removes a given player from the lobby
func (lobby *Lobby) RemovePlayer(player *player.Player) error {
	lobby.Lock()
	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).Delete(&LobbySlot{}).Error
	lobby.Unlock()

	if err != nil {
		return err
	}

	rpc.DisallowPlayer(lobby.ID, player.SteamID, player.ID)
	lobby.OnChange(true)
	return nil
}

//BanPlayer bans a given player from the lobby
func (lobby *Lobby) BanPlayer(player *player.Player) {
	rpc.DisallowPlayer(lobby.ID, player.SteamID, player.ID)
	db.DB.Model(lobby).Association("BannedPlayers").Append(player)
}

//ReadyPlayer readies up given player, use when lobby.State == LobbyStateWaiting
func (lobby *Lobby) ReadyPlayer(player *player.Player) error {
	err := db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("ready", true).Error
	if err != nil {
		return errors.New("Player is not in the lobby.")
	}
	lobby.OnChange(false)
	return nil
}

//UnreadyPlayer unreadies given player, use when lobby.State == LobbyStateWaiting
func (lobby *Lobby) UnreadyPlayer(player *player.Player) error {
	err := db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("ready", false).Error
	if err != nil {
		return errors.New("Player is not in the lobby.")
	}

	lobby.OnChange(false)
	return nil
}

//GetUnreadyPlayers returns a list of unready players in the lobby.
//only used when lobby state == LobbyStateReadyingUp
func (lobby *Lobby) GetUnreadyPlayers() (players []*player.Player) {
	db.DB.Model(&player.Player{}).Joins("INNER JOIN lobby_slots ON lobby_slots.player_id = players.id").Where("lobby_slots.lobby_id = ? AND lobby_slots.ready = ?", lobby.ID, false).Find(&players)
	return
}

//RemoveUnreadyPlayers removes players who haven't removed. If spec == true, move them to spectators
func (lobby *Lobby) RemoveUnreadyPlayers(spec bool) error {
	playerids := []uint{}

	if spec {
		//get list of player ids which are not ready
		err := db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND ready = ?", lobby.ID, false).Pluck("player_id", &playerids).Error
		if err != nil {
			return err
		}
	}

	//remove players which aren't ready
	lobby.Lock()
	err := db.DB.Where("lobby_id = ? AND ready = ?", lobby.ID, false).Delete(&LobbySlot{}).Error
	lobby.Unlock()

	if spec {
		for _, id := range playerids {
			p, _ := player.GetPlayerByID(id)
			lobby.AddSpectator(p)
		}
	}
	lobby.OnChange(true)
	return err
}

var (
	inGameMu    = new(sync.Mutex)
	inGameTimer = make(map[uint]*time.Timer)
)

//AfterPlayerNotInGameFunc waits the duration to elapse, and if the given player
//is still not in the game server, calls f in it's own goroutine.
func (lobby *Lobby) AfterPlayerNotInGameFunc(player *player.Player, d time.Duration, f func()) {
	helpers.GlobalWait.Add(1)

	inGameMu.Lock()
	inGameTimer[player.ID] = time.AfterFunc(d, func() {
		var count int
		db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND player_id = ? AND needs_sub = FALSE AND in_game = FALSE", lobby.ID, player.ID).Count(&count)

		if count != 0 && lobby.CurrentState() != Ended {
			f()
		}
		helpers.GlobalWait.Done()
	})
	inGameMu.Unlock()
}

//IsPlayerInGame returns true if the player is in-game
func (lobby *Lobby) IsPlayerInGame(player *player.Player) bool {
	var count int
	db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND player_id = ? AND in_game = TRUE", lobby.ID, player.ID).Count(&count)
	return count != 0
}

func (lobby *Lobby) IsPlayerInMumble(player *player.Player) bool {
	var count int
	db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND player_id = ? AND in_mumble = TRUE", lobby.ID, player.ID).Count(&count)
	return count != 0
}

//IsPlayerReady returns true if the given player is ready
func (lobby *Lobby) IsPlayerReady(player *player.Player) (bool, error) {
	var ready bool
	err := db.DB.DB().QueryRow("SELECT ready FROM lobby_slots WHERE lobby_id = $1 AND player_id = $2", lobby.ID, player.ID).Scan(&ready)
	return ready, err
}

//UnreadyAllPlayers unreadies all players in the lobby
func (lobby *Lobby) UnreadyAllPlayers() error {
	lobby.Lock()
	err := db.DB.Model(&LobbySlot{}).Where("lobby_id = ?", lobby.ID).UpdateColumn("ready", false).Error
	lobby.Unlock()

	lobby.OnChange(false)
	return err
}

//ReadyUpTimeLeft returns the amount of time left to ready up before a player is kicked.
//Use when a player reconnects while lobby.State == LobbyStateReadyingUp and the player isn't ready
func (lobby *Lobby) ReadyUpTimeLeft() int64 {
	return int64(lobby.ReadyUpTimestamp - time.Now().Unix())
}

//IsEveryoneReady returns whether all players have readied up.
func (lobby *Lobby) IsEveryoneReady() bool {
	readyPlayers := 0
	db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND ready = ?", lobby.ID, true).Count(&readyPlayers)

	return readyPlayers == 2*format.NumberOfClassesMap[lobby.Type]
}

//AddSpectator adds a given player as a lobby spectator
func (lobby *Lobby) AddSpectator(player *player.Player) error {
	err := db.DB.Model(lobby).Association("Spectators").Append(player).Error
	if err != nil {
		return err
	}
	lobby.OnChange(false)
	return nil
}

//RemoveSpectator removes the given player from the lobby spectators list. If broadcast, then
//broadcast the change to other players
func (lobby *Lobby) RemoveSpectator(player *player.Player, broadcast bool) error {
	err := db.DB.Model(lobby).Association("Spectators").Delete(player).Error
	if err != nil {
		return err
	}
	if broadcast {
		lobby.OnChange(false)
	}
	return nil
}

//GetPlayerNumber returns the number of occupied slots in the lobby
func (lobby *Lobby) GetPlayerNumber() int {
	count := 0
	err := db.DB.Model(&LobbySlot{}).Where("lobby_id = ?", lobby.ID).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

//IsFull returns whether all lobby spots have been filled
func (lobby *Lobby) IsFull() bool {
	return lobby.GetPlayerNumber() == 2*format.NumberOfClassesMap[lobby.Type]
}

//IsSlotFilled returns whether the given slot (by number) is occupied by a player
func (lobby *Lobby) IsSlotFilled(slot int) bool {
	_, err := lobby.GetPlayerIDBySlot(slot)
	if err != nil {
		return false
	}
	return true
}

//GetAllSlots returns a list of all occupied slots in the lobby
func (lobby *Lobby) GetAllSlots() []LobbySlot {
	db.DB.Preload("Slots").First(lobby, lobby.ID)
	return lobby.Slots
}

func (lobby *Lobby) DiscordNotif(msg string) {
	if helpers.Discord != nil {
		mumble := ""
		if lobby.Mumble {
			mumble = helpers.DiscordEmoji("mumble")
		}

		region := lobby.RegionName
		if lobby.RegionCode == "eu" || lobby.RegionCode == "au" {
			region = fmt.Sprintf(":flag_%s:", lobby.RegionCode)
		} else if lobby.RegionCode == "na" {
			region = ":flag_us:"
		}

		byLine := ""
		player, playerErr := player.GetPlayerBySteamID(lobby.CreatedBySteamID)
		if playerErr != nil {
			logrus.Error(playerErr)
		} else {
			byLine = fmt.Sprintf(" by %s", player.Alias())
		}

		formatName := format.FriendlyNamesMap[lobby.Type]

		msg := fmt.Sprintf("%s%s%s lobby on %s%s: %s %s/lobby/%d", region, mumble, formatName, lobby.MapName, byLine, msg, config.Constants.LoginRedirectPath, lobby.ID)
		helpers.DiscordSendToChannel("lobby-notifications", msg)
		helpers.DiscordSendToChannel(fmt.Sprintf("%s-%s", formatName, lobby.RegionCode), msg)

		msg = fmt.Sprintf("@here %s", msg)
		helpers.DiscordSendToChannel("lobby-notifications-ping", msg)
		helpers.DiscordSendToChannel(fmt.Sprintf("%s-%s-ping", formatName, lobby.RegionCode), msg)
	}
}

//SetupServer setups the TF2 server for the lobby, creates the mumble channels for it
func (lobby *Lobby) SetupServer() error {
	if lobby.State == Ended {
		return errors.New("Lobby is closed")
	}

	err := rpc.SetupServer(lobby.ID, lobby.ServerInfo, lobby.Type, lobby.League, lobby.Whitelist, lobby.MapName)
	if err != nil {
		return err
	}

	rpc.FumbleLobbyCreated(lobby.ID)
	lobby.DiscordNotif("Join")
	return nil
}

//Close closes the lobby, which has the following effects:
//
//  All unfilled substitutes for the lobby are "filled" (ie, their filled field is set to true)
//  The corresponding ServerRecord is deleted
//
//If rpc == true, the log listener in Pauling for the corresponding server is stopped, this is
//used when the lobby is closed manually by a player
func (lobby *Lobby) Close(doRPC, matchEnded bool) {
	var count int

	db.DB.Preload("ServerInfo").First(lobby, lobby.ID)
	db.DB.Model(&gameserver.ServerRecord{}).Where("host = ?", lobby.ServerInfo.Host).Count(&count)
	if count != 0 {
		gameserver.PutStoredServer(lobby.ServerInfo.Host)
	}

	lobby.SetState(Ended)
	db.DB.First(lobby).UpdateColumn("match_ended", matchEnded)
	//db.DB.Exec("DELETE FROM spectators_players_lobbies WHERE lobby_id = ?", lobby.ID)
	if doRPC {
		rpc.End(lobby.ID)
	}
	if matchEnded {
		lobby.UpdateStats()
	}
	if lobby.ServemeID != 0 {
		context := helpers.GetServemeContext(lobby.ServerInfo.Host)
		err := context.Delete(lobby.ServemeID, lobby.CreatedBySteamID)
		if err != nil {
			logrus.Error(err)
		}
		if matchEnded {
			time.AfterFunc(10*time.Second, func() {
				lobby.DownloadDemo(context)
			})
		}
	}

	privateRoom := fmt.Sprintf("%d_private", lobby.ID)
	broadcaster.SendMessageToRoom(privateRoom, "lobbyLeft", LobbyEvent{ID: lobby.ID})

	publicRoom := fmt.Sprintf("%d_public", lobby.ID)
	broadcaster.SendMessageToRoom(publicRoom, "lobbyClosed", DecorateLobbyClosed(lobby))

	db.DB.Model(&gameserver.ServerRecord{}).Where("id = ?", lobby.ServerInfoID).Delete(&gameserver.ServerRecord{})
	BroadcastSubList()
	BroadcastLobby(lobby)
	BroadcastLobbyList() // has to be done manually for now
	rpc.FumbleLobbyEnded(lobby.ID)
	lobby.deleteLock()
}

func (lobby *Lobby) DownloadDemo(context *servemetf.Context) {
	file := fmt.Sprintf("%s/%d.dem", config.Constants.DemosFolder,
		lobby.ID)
	err := context.DownloadDemo(lobby.ServemeID, lobby.CreatedBySteamID, file)
	if err != nil {
		logrus.Error(err)
	} else {
		url := fmt.Sprintf("%s/demos/%d.dem", config.Constants.PublicAddress, lobby.ID)
		chat.SendNotification("STV Demo for this lobby is available at "+url, int(lobby.ID))
	}
}

//UpdateStats updates the PlayerStats records for all players in the lobby
//(increments the relevent lobby type field by one). Used when the lobby successfully ends.
func (lobby *Lobby) UpdateStats() {
	db.DB.Preload("Slots").First(lobby, lobby.ID)

	for _, slot := range lobby.Slots {
		p, _ := player.GetPlayerByID(slot.PlayerID)

		db.DB.Preload("Stats").First(p, slot.PlayerID)
		p.Stats.PlayedCountIncrease(lobby.Type)
		p.Stats.IncreaseClassCount(lobby.Type, slot.Slot)
		p.Save()
	}
	lobby.OnChange(false)
}

func (lobby *Lobby) UpdateHours(logsID int) error {
	db.DB.Model(&Lobby{}).Where("id = ?", lobby.ID).UpdateColumn("logstf_id", logsID)

	logs, err := logstf.GetLogs(logsID)
	if err != nil {
		return err
	}

	for steamID, playerStats := range logs.Players {
		commid, _ := steamid.SteamIdToCommId(steamID)
		player, err := player.GetPlayerWithStats(commid)
		if err != nil {
			logrus.Error("Couldn't find player with SteamID ", commid)
			continue
		}

		for _, class := range playerStats.ClassStats {
			totalTime := time.Second * time.Duration(class.TotalTime)

			switch class.Type {
			case "scout":
				player.Stats.ScoutHours += totalTime
			case "soldier":
				player.Stats.SoldierHours += totalTime
			case "demoman":
				player.Stats.DemoHours += totalTime
			case "heavyweapons":
				player.Stats.HeavyHours += totalTime
			case "pyro":
				player.Stats.PyroHours += totalTime
			case "engineer":
				player.Stats.EngineerHours += totalTime
			case "spy":
				player.Stats.SpyHours += totalTime
			case "sniper":
				player.Stats.SniperHours += totalTime
			case "medic":
				player.Stats.MedicHours += totalTime
			}
		}

		player.Stats.Save()
	}

	return nil
}

func (lobby *Lobby) setInGameStatus(player *player.Player, inGame bool) error {
	err := db.DB.Model(&LobbySlot{}).Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).UpdateColumn("in_game", inGame).Error

	lobby.OnChange(false)
	return err
}

//SetInGame sets the in-game status of the given player to true
func (lobby *Lobby) SetInGame(player *player.Player) error {
	inGameMu.Lock()
	timer, ok := inGameTimer[player.ID]
	if ok {
		if timer.Stop() {
			helpers.GlobalWait.Done()
		}
		delete(inGameTimer, player.ID)
	}
	inGameMu.Unlock()

	return lobby.setInGameStatus(player, true)
}

//SetNotInGame sets the in-game status of the given player to false
func (lobby *Lobby) SetNotInGame(player *player.Player) error {
	return lobby.setInGameStatus(player, false)
}

func (lobby *Lobby) setInMumbleStatus(player *player.Player, inMumble bool) error {
	err := db.DB.Model(&LobbySlot{}).Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).UpdateColumn("in_mumble", inMumble).Error

	lobby.OnChange(false)
	return err
}

//SetInMumble sets the in-mumble status of the given player to true
func (lobby *Lobby) SetInMumble(player *player.Player) error {
	return lobby.setInMumbleStatus(player, true)
}

//SetNotInMumble sets the in-mumble status of the given player to false
func (lobby *Lobby) SetNotInMumble(player *player.Player) error {
	return lobby.setInMumbleStatus(player, false)
}

//Start sets lobby.State to LobbyStateInProgress, calls SubNotInGamePlayers after 5 minutes
func (lobby *Lobby) Start() {
	rows := db.DB.Model(&Lobby{}).Where("id = ? AND state <> ?", lobby.ID, InProgress).Update("state", InProgress).RowsAffected
	if rows != 0 { // if == 0, then game is already in progress
		go rpc.ReExecConfig(lobby.ID, false)

		// var playerids []uint
		// db.DB.Model(&LobbySlot{}).Where("lobby_id = ?", lobby.ID).Pluck("player_id", &playerids)

		// for _, id := range playerids {
		// 	player, _ := GetPlayerByID(id)
		// 	lobby.AfterPlayerNotInGameFunc(player, 5*time.Minute, func() {
		// 		lobby.Substitute(player)
		// 	})
		// }
	}
}

//OnChange broadcasts the given lobby to other players. If base is true, broadcasts the lobby list too.
func (lobby *Lobby) OnChange(base bool) {
	switch lobby.State {
	case Waiting, InProgress, ReadyingUp:
		BroadcastLobby(lobby)
		if base {
			BroadcastLobbyList()
		}
	}
}

//BroadcastLobby broadcasts the lobby to the lobby's public room (id_public)
func BroadcastLobby(lobby *Lobby) {
	room := strconv.FormatUint(uint64(lobby.ID), 10)

	broadcaster.SendMessageToRoom(fmt.Sprintf("%s_public", room), "lobbyData", DecorateLobbyData(lobby, true))
}

//BroadcastLobbyToUser broadcasts the lobby to the a user with the given steamID
func BroadcastLobbyToUser(lobby *Lobby, steamid string) {
	broadcaster.SendMessage(steamid, "lobbyData", DecorateLobbyData(lobby, true))
}

//BroadcastLobbyList broadcasts the lobby list to all users
func BroadcastLobbyList() {
	broadcaster.SendMessageToRoom(
		"0_public",
		"lobbyListData", DecorateLobbyListData(GetWaitingLobbies(), false))
}

var maxSubs = map[format.Format]int{
	format.Highlander: 5,
	format.Sixes:      4,
	format.Bball:      2,
	format.Ultiduo:    2,
	format.Debug:      2,
	format.Fours:      2,
}

//Substitute sets the needs_sub column of the given slot to true, and broadcasts the new
//substitute list
func (lobby *Lobby) Substitute(player *player.Player) {
	lobby.Lock()
	db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("needs_sub", true)
	lobby.Unlock()

	var count int
	db.DB.Model(&LobbySlot{}).Where("lobby_id = ? AND needs_sub = TRUE", lobby.ID).Count(&count)
	if count == maxSubs[lobby.Type] {
		chat.SendNotification("Lobby closed (Too many subs).", int(lobby.ID))
		lobby.Close(true, false)
	}

	db.DB.Preload("Stats").First(player, player.ID)
	player.Stats.IncreaseSubCount()
	BroadcastSubList()
}

//BroadcastSubList broadcasts a the subtitute list to the room 0_public
func BroadcastSubList() {
	subList := DecorateSubstituteList()
	broadcaster.SendMessageToRoom("0_public", "subListData", subList)
}

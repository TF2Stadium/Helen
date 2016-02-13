// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/jinzhu/gorm"
)

type LobbyType int
type LobbyState int

const (
	LobbyTypeSixes LobbyType = iota
	LobbyTypeHighlander
	LobbyTypeFours
	LobbyTypeUltiduo
	LobbyTypeBball
	LobbyTypeDebug

	LobbyStateInitializing LobbyState = 0
	LobbyStateWaiting      LobbyState = 1
	LobbyStateReadyingUp   LobbyState = 2
	LobbyStateInProgress   LobbyState = 3
	LobbyStateEnded        LobbyState = 5
)

var (
	stateString = map[LobbyState]string{
		LobbyStateWaiting:    "Waiting For Players",
		LobbyStateInProgress: "Lobby in Progress",
		LobbyStateEnded:      "Lobby Ended",
	}

	formatMap = map[LobbyType]string{
		LobbyTypeSixes:      "6s",
		LobbyTypeHighlander: "Highlander",
		LobbyTypeFours:      "4v4",
		LobbyTypeUltiduo:    "Ultiduo",
		LobbyTypeBball:      "Bball",
		LobbyTypeDebug:      "Debug",
	}
)
var (
	LobbyBanErr        = helpers.NewTPError("You have been banned from this lobby", 4)
	BadSlotErr         = helpers.NewTPError("This slot does not exist", 3)
	FilledErr          = helpers.NewTPError("This slot has been filled", 2)
	NotWhitelistedErr  = helpers.NewTPError("You aren't allowed in this lobby", 3)
	InvalidPasswordErr = helpers.NewTPError("Invalid slot password", 3)
	NeedsSubErr        = helpers.NewTPError("This slot needs a substitute", 3)

	ReqHoursErr       = helpers.NewTPError("You don't have sufficient hours to join this lobby", 3)
	ReqLobbiesErr     = helpers.NewTPError("You haven't played sufficient lobbies to join this lobby", 3)
	ReqReliabilityErr = helpers.NewTPError("You have insufficient reliability to join this lobby", 3)
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
	NeedsSub bool //true if the slot needs a subtitute player
}

type ServerRecord struct {
	ID             uint
	Host           string
	LogSecret      string
	ServerPassword string // sv_password
	RconPassword   string // rcon_password
}

//DeleteUnusedServerRecords checks all server records in the DB and deletes them if
//the corresponsing lobby is closed
func DeleteUnusedServerRecords() {
	serverInfoIDs := []uint{}
	db.DB.Table("server_records").Pluck("id", &serverInfoIDs)
	for _, id := range serverInfoIDs {
		lobby := &Lobby{}
		err := db.DB.Where("server_info_id = ?", id).First(lobby).Error

		if err != nil || lobby.State == LobbyStateEnded {
			db.DB.Table("server_records").Where("id = ?", id).Delete(&ServerRecord{})
		}
	}
}

//Lobby represents a Lobby
type Lobby struct {
	gorm.Model
	State LobbyState

	Mode    string    // Game Mode
	MapName string    // Map Name
	Type    LobbyType // League config used
	League  string

	RegionCode string // Region Code ("na", "eu", etc)
	RegionName string // Region Name ("North America", "Europe", etc)

	Mumble bool // Whether mumble is required

	Slots []LobbySlot // List of occupied slots

	SlotPassword    string // Slot password, if any
	PlayerWhitelist string // URL of steam group
	TwitchChannel   string // twitch channel, slots will be restricted to channel subs

	// TF2 Server Info
	ServerInfo   ServerRecord
	ServerInfoID uint

	Whitelist string //whitelist.tf ID

	Spectators    []Player `gorm:"many2many:spectators_players_lobbies"` // List of spectators
	BannedPlayers []Player `gorm:"many2many:banned_players_lobbies"`     // List of Banned Players

	CreatedBySteamID string // SteamID of the lobby leader/creator

	ReadyUpTimestamp int64 // (Unix) Timestamp at which the ready up timeout started
}

// Requirement stores a requirement for a particular slot in a lobby
type Requirement struct {
	ID      uint `json:"-"`
	LobbyID uint `json:"-"`

	Slot int `json:"-"` // if -1, applies to all slots

	Hours       int     `json:"hours"`       // minimum hours needed
	Lobbies     int     `json:"lobbies"`     // minimum lobbies played
	Reliability float64 `json:"reliability"` // minimum reliability needed
}

func NewRequirement(lobbyID uint, slot int, hours int, lobbies int) *Requirement {
	r := &Requirement{
		LobbyID: lobbyID,
		Slot:    slot,
		Hours:   hours,
		Lobbies: lobbies}
	db.DB.Save(r)

	return r
}

func (r *Requirement) Save() { db.DB.Save(r) }

//GetSlotRequirement returns the slot requirement for the lobby lobby
func (lobby *Lobby) GetSlotRequirement(slot int) (*Requirement, error) {
	req := &Requirement{}
	err := db.DB.Table("requirements").Where("lobby_id = ? AND slot = ?", lobby.ID, slot).First(req).Error

	return req, err
}

//HasSlotRequirement returns true if the given slot in the lobby has a requirement
func (lobby *Lobby) HasSlotRequirement(slot int) bool {
	var count int
	db.DB.Table("requirements").Where("lobby_id = ? AND slot = ?", lobby.ID, slot).Count(&count)
	return count != 0
}

//HasRequirements returns true if the given slot has a requirement (either general or slot-only)
func (lobby *Lobby) HasRequirements(slot int) bool {
	return lobby.HasSlotRequirement(slot)
}

//FitsRequirements checks if the player fits the requirement to be added to the given slot in the lobby
func (l *Lobby) FitsRequirements(player *Player, slot int) (bool, *helpers.TPError) {
	//BUG(vibhavp): FitsRequirements doesn't check reliability
	var req *Requirement

	slotReq, err := l.GetSlotRequirement(slot)
	if err == nil {
		req = slotReq
	}

	db.DB.Preload("Stats").First(player, player.ID)

	if time.Since(player.UpdatedAt) < time.Hour*time.Duration(req.Hours-player.GameHours) {
		//update player info only if the number of hours needed > the number of hours
		//passed since player info was last updated
		player.UpdatePlayerInfo()
		player.Save()
	}

	if player.GameHours < req.Hours {
		return false, ReqHoursErr
	}

	if player.Stats.TotalLobbies() < req.Lobbies {
		return false, ReqLobbiesErr
	}

	return true, nil
}

func getGamemode(mapName string, lobbyType LobbyType) string {
	switch {
	case strings.HasPrefix(mapName, "koth"):
		if lobbyType == LobbyTypeUltiduo {
			return "ultiduo"
		}

		return "koth"

	case strings.HasPrefix(mapName, "ctf"):
		if lobbyType == LobbyTypeBball {
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

// Returns a new lobby object with the given parameters
func NewLobby(mapName string, lobbyType LobbyType, league string, serverInfo ServerRecord, whitelist string, mumble bool, whitelistGroup, password string) *Lobby {
	lobby := &Lobby{
		Mode:            getGamemode(mapName, lobbyType),
		Type:            lobbyType,
		State:           LobbyStateInitializing,
		League:          league,
		MapName:         mapName,
		Whitelist:       whitelist, // that's a strange line
		Mumble:          mumble,
		ServerInfo:      serverInfo,
		PlayerWhitelist: whitelistGroup,
		SlotPassword:    password,
	}

	// Must specify CreatedBy manually if the lobby is created by a player

	return lobby
}

//Delete removes the lobby object from the database.
//Closed lobbies aren't deleted, this function is used for
//lobbies where the game server had an error while being setup.
func (lobby *Lobby) Delete() {
	db.DB.Delete(lobby)
	db.DB.Delete(&lobby.ServerInfo)
}

//GetWaitingLobbies returns a list of lobby objects that haven't been filled yet
func GetWaitingLobbies() (lobbies []*Lobby) {
	db.DB.Where("state = ?", LobbyStateWaiting).Order("id desc").Find(&lobbies)
	return
}

//CurrentState returns the lobby's current state.
//It's meant to be used for old lobby objects which might have their state change while the
//object hasn't been updated.
func (l *Lobby) CurrentState() LobbyState {
	var state int
	db.DB.DB().QueryRow("SELECT state FROM lobbies WHERE id = $1", l.ID).Scan(&state)
	return LobbyState(state)
}

func (l *Lobby) SetState(s LobbyState) {
	db.DB.Table("lobbies").Where("id = ?", l.ID).UpdateColumn("state", s)
}

//GetPlayerSlotObj returns the LobbySlot object if the given player occupies a slot in the lobby.
func (lobby *Lobby) GetPlayerSlotObj(player *Player) (*LobbySlot, error) {
	slotObj := &LobbySlot{}

	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).First(slotObj).Error

	return slotObj, err
}

//GetPlayerSlot returns the slot number if the player occupies a slot int eh lobby
func (lobby *Lobby) GetPlayerSlot(player *Player) (int, error) {
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

	lobby.RealAfterSave()
	return err
}

//GetLobbyByIdServer returns the lobby object, plus the ServerInfo object inside it
func GetLobbyByIDServer(id uint) (*Lobby, *helpers.TPError) {
	nonExistentLobby := helpers.NewTPError("Lobby not in the database", -1)

	lob := &Lobby{}
	err := db.DB.Preload("ServerInfo").First(lob, id).Error

	if err != nil {
		return nil, nonExistentLobby
	}

	return lob, nil
}

//GetLobbyByID returns lobby object, without the ServerInfo object inside it.
func GetLobbyByID(id uint) (*Lobby, *helpers.TPError) {
	nonExistentLobby := helpers.NewTPError("Lobby not in the database", -1)

	lob := &Lobby{}
	err := db.DB.First(lob, id).Error

	if err != nil {
		return nil, nonExistentLobby
	}

	return lob, nil
}

//HasPlayer returns true if the given player occupies a slot in the lobby
func (lobby *Lobby) HasPlayer(player *Player) bool {
	var count int
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).Count(&count)

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
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND slot = ?", lobby.ID, slot).UpdateColumn("needs_sub", false).Error
	BroadcastSubList()
	return err
}

//AddPlayer adds the given player to lobby, If the player occupies a slot in the lobby already, switch slots.
//If the player is in another lobby, removes them from that lobby before adding them.
func (lobby *Lobby) AddPlayer(player *Player, slot int, password string) *helpers.TPError {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */

	//check if slot password is valid
	if lobby.SlotPassword != "" && lobby.SlotPassword != password {
		return InvalidPasswordErr
	}

	num := 0

	//Check if player is banned
	//TODO(nonagon): It should really be possible to do this query using relations
	if err := db.DB.Table("banned_players_lobbies").
		Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).
		Count(&num).Error; num > 0 || err != nil {
		//logrus.Debug(fmt.Sprint(err))
		return LobbyBanErr
	}

	if slot >= 2*NumberOfClassesMap[lobby.Type] || slot < 0 {
		return BadSlotErr
	}

	var slotChange bool
	//Check if the player is currently in another lobby
	if currLobbyID, err := player.GetLobbyID(false); err == nil {
		if currLobbyID != lobby.ID {
			//if the player is in a different lobby, remove them from that lobby
			//plus substitute them
			curLobby, _ := GetLobbyByID(currLobbyID)

			if curLobby.State == LobbyStateInProgress {
				curLobby.Substitute(player)
			} else {
				curLobby.RemovePlayer(player)
				curLobby.AddSpectator(player)
			}

		} else { //player is in the same lobby, they're changing their slots
			//assign the player to a new slot
			if lobby.SlotNeedsSubstitute(slot) {
				//the slot needs a substitute (which happens when the lobby is in progress),
				//so players already in the lobby cannot fill it.
				return NeedsSubErr
			}
			db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).Delete(&LobbySlot{})
			slotChange = true
		}
	}

	if !slotChange {
		//check if the player is in the steam group whitelist
		url := fmt.Sprintf(`http://steamcommunity.com/groups/%s/memberslistxml/?xml=1`,
			lobby.PlayerWhitelist)

		if lobby.PlayerWhitelist != "" && !helpers.IsWhitelisted(player.SteamID, url) {
			return NotWhitelistedErr
		}

		if lobby.HasRequirements(slot) {
			//check if player fits the requirements for the slot
			if ok, err := lobby.FitsRequirements(player, slot); !ok {
				return err
			}
		}

		//check if player has been subbed to the twitch channel (if any)
		if lobby.TwitchChannel != "" {
			//check if player has connected their twitch account
			if player.TwitchAccessToken == "" {
				return helpers.NewTPError("You need to connect your Twitch Account first to join the lobby.", -1)
			}
			if !player.IsSubscribed(lobby.TwitchChannel) {
				err := fmt.Sprintf("You aren't subscribed to %s", lobby.TwitchChannel)
				return helpers.NewTPError(err, -1)
			}
		}
	}

	// Check if player is a substitute (the slot needs a subtitute)
	if lobby.SlotNeedsSubstitute(slot) {
		//kicks previous slot occupant if they're in-game, resets their !rep count, removes them from the lobby
		DisallowPlayer(lobby.ID, player.SteamID)
		//delete previous slot
		db.DB.Where("lobby_id = ? AND slot = ?", lobby.ID, slot).Delete(&LobbySlot{})
		BroadcastSubList() //since the sub slot has been deleted, broadcast the updated substitute list
		//notify players in game server of subtitute
		class, team, _ := LobbyGetSlotInfoString(lobby.Type, slot)
		Say(lobby.ID, fmt.Sprintf("Substitute found for %s %s: %s (%s)", team, class, player.Name, player.SteamID))
		//allow player in mumble
		FumbleLobbyPlayerJoinedSub(lobby, player, slot)
	} else if _, err := lobby.GetPlayerIDBySlot(slot); err == nil {
		return FilledErr
	} else {
		FumbleLobbyPlayerJoined(lobby, player, slot) // no errors, al
	}

	//try to remove them from spectators
	lobby.RemoveSpectator(player, true)

	newSlotObj := &LobbySlot{
		PlayerID: player.ID,
		LobbyID:  lobby.ID,
		Slot:     slot,
	}

	db.DB.Create(newSlotObj)

	lobby.OnChange(true)

	return nil
}

//RemovePlayer removes a given player from the lobby
func (lobby *Lobby) RemovePlayer(player *Player) *helpers.TPError {
	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).Delete(&LobbySlot{}).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}

	DisallowPlayer(lobby.ID, player.SteamID)
	lobby.OnChange(true)
	return nil
}

//BanPlayer bans a given player from the lobby
func (lobby *Lobby) BanPlayer(player *Player) {
	DisallowPlayer(lobby.ID, player.SteamID)
	db.DB.Model(lobby).Association("BannedPlayers").Append(player)
}

//ReadyPlayer readies up given player, use when lobby.State == LobbyStateWaiting
func (lobby *Lobby) ReadyPlayer(player *Player) *helpers.TPError {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("ready", true).Error
	if err != nil {
		return helpers.NewTPError("Player is not in the lobby.", 5)
	}
	lobby.OnChange(false)
	return nil
}

//UnreadyPlayer unreadies given player, use when lobby.State == LobbyStateWaiting
func (lobby *Lobby) UnreadyPlayer(player *Player) *helpers.TPError {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("ready", false).Error
	if err != nil {
		return helpers.NewTPError("Player is not in the lobby.", 5)
	}

	lobby.OnChange(false)
	return nil
}

//GetUnreadyPlayers returns a list of unready players in the lobby.
//only used when lobby state == LobbyStateReadyingUp
func (lobby *Lobby) GetUnreadyPlayers() (players []*Player) {
	db.DB.Table("players").Joins("INNER JOIN lobby_slots ON lobby_slots.player_id = players.id").Where("lobby_slots.lobby_id = ? AND lobby_slots.ready = ?", lobby.ID, false).Find(&players)
	return
}

//RemoveUnreadyPlayers removes players who haven't removed. If spec == true, move them to spectators
func (lobby *Lobby) RemoveUnreadyPlayers(spec bool) error {
	playerids := []uint{}

	if spec {
		//get list of player ids which are not ready
		err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND ready = ?", lobby.ID, false).Pluck("player_id", &playerids).Error
		if err != nil {
			return err
		}
	}

	//remove players which aren't ready
	err := db.DB.Where("lobby_id = ? AND ready = ?", lobby.ID, false).Delete(&LobbySlot{}).Error
	if spec {
		for _, id := range playerids {
			player := &Player{}
			db.DB.First(player, id)
			lobby.AddSpectator(player)
		}
	}
	lobby.OnChange(true)
	return err
}

//IsPlayerInGame returns true if the player is in-game
func (lobby *Lobby) IsPlayerInGame(player *Player) (bool, error) {
	var ingame bool
	err := db.DB.DB().QueryRow("SELECT in_game FROM lobby_slots WHERE lobby_id = $1 AND player_id = $2", lobby.ID, player.ID).Scan(&ingame)
	if err != nil {
		return false, err
	}

	return ingame, err
}

//IsPlayerReady returns true if the given player is ready
func (lobby *Lobby) IsPlayerReady(player *Player) (bool, error) {
	var ready bool
	err := db.DB.DB().QueryRow("SELECT ready FROM lobby_slots WHERE lobby_id = $1 AND player_id = $2", lobby.ID, player.ID).Scan(&ready)
	return ready, err
}

//UnreadyAllPlayers unreadies all players in the lobby
func (lobby *Lobby) UnreadyAllPlayers() error {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).UpdateColumn("ready", false).Error

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
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND ready = ?", lobby.ID, true).Count(&readyPlayers)

	return readyPlayers == 2*NumberOfClassesMap[lobby.Type]
}

//AddSpectator adds a given player as a lobby spectator
func (lobby *Lobby) AddSpectator(player *Player) *helpers.TPError {
	err := db.DB.Model(lobby).Association("Spectators").Append(player).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}
	lobby.OnChange(false)
	return nil
}

//RemoveSpectator removes the given player from the lobby spectators list. If broadcast, then
//broadcast the change to other players
func (lobby *Lobby) RemoveSpectator(player *Player, broadcast bool) *helpers.TPError {
	err := db.DB.Model(lobby).Association("Spectators").Delete(player).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}
	if broadcast {
		lobby.OnChange(false)
	}
	return nil
}

//GetPlayerNumber returns the number of occupied slots in the lobby
func (lobby *Lobby) GetPlayerNumber() int {
	count := 0
	err := db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

//IsFull returns whether all lobby spots have been filled
func (lobby *Lobby) IsFull() bool {
	return lobby.GetPlayerNumber() == 2*NumberOfClassesMap[lobby.Type]
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

//SetupServer setups the TF2 server for the lobby, calls Pauling.SetupServer()
func (lobby *Lobby) SetupServer() error {
	if lobby.State == LobbyStateEnded {
		return nil
	}

	err := SetupServer(lobby.ID, lobby.ServerInfo, lobby.Type, lobby.League, lobby.Whitelist, lobby.MapName)
	return err
}

//Close closes the lobby, which has the following effects:
//
//  All unfilled substitutes for the lobby are "filled" (ie, their filled field is set to true)
//  The corresponding ServerRecord is deleted
//
//If rpc == true, the log listener in Pauling for the corresponding server is stopped, this is
//used when the lobby is closed manually by a player
func (lobby *Lobby) Close(rpc bool) {
	db.DB.First(&lobby).UpdateColumn("state", LobbyStateEnded)
	db.DB.Table("server_records").Where("id = ?", lobby.ServerInfoID).Delete(&ServerRecord{})
	db.DB.Table("requirements").Where("lobby_id = ?", lobby.ID).Delete(&Requirement{})
	//db.DB.Exec("DELETE FROM spectators_players_lobbies WHERE lobby_id = ?", lobby.ID)
	if rpc {
		End(lobby.ID)
	}

	privateRoom := fmt.Sprintf("%d_private", lobby.ID)
	broadcaster.SendMessageToRoom(privateRoom, "lobbyLeft", DecorateLobbyLeave(lobby))

	publicRoom := fmt.Sprintf("%d_public", lobby.ID)
	broadcaster.SendMessageToRoom(publicRoom, "lobbyClosed", DecorateLobbyClosed(lobby))

	BroadcastSubList()
	BroadcastLobby(lobby)
	BroadcastLobbyList() // has to be done manually for now
	FumbleLobbyEnded(lobby)
}

//UpdateStats updates the PlayerStats records for all players in the lobby
//(increments the relevent lobby type field by one). Used when the lobby successfully ends.
func (lobby *Lobby) UpdateStats() {
	db.DB.Preload("Slots").First(lobby, lobby.ID)

	for _, slot := range lobby.Slots {
		player := &Player{}
		err := db.DB.First(player, slot.PlayerID).Error
		if err != nil {
			logrus.Error(err)
			return
		}
		db.DB.Preload("Stats").First(player, slot.PlayerID)
		player.Stats.PlayedCountIncrease(lobby.Type)
		player.Stats.IncreaseClassCount(lobby, slot.Slot)
		player.Save()
	}
	lobby.OnChange(false)
}

func (lobby *Lobby) setInGameStatus(player *Player, inGame bool) error {
	err := db.DB.Table("lobby_slots").Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).UpdateColumn("in_game", inGame).Error

	lobby.OnChange(false)
	return err
}

//SetInGame sets the in-game status of the given player to true
func (lobby *Lobby) SetInGame(player *Player) error {
	return lobby.setInGameStatus(player, true)
}

//SetNotInGame sets the in-game status of the given player to false
func (lobby *Lobby) SetNotInGame(player *Player) error {
	return lobby.setInGameStatus(player, false)
}

//SubNotInGamePlayers substitutes players who haven't joined the game server yet.
//Called 5 minutes after the lobby starts.
func (lobby *Lobby) SubNotInGamePlayers() {
	playerids := []uint{}
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND in_game = ?", lobby.ID, false).Pluck("player_id", &playerids)
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND in_game = ? AND needs_sub = ?", lobby.ID, false, false).UpdateColumn("needs_sub", true)

	for _, id := range playerids {
		player, _ := GetPlayerByID(id)
		SendNotification(fmt.Sprintf("%s has been reported for not joining the game.", player.Alias()), int(lobby.ID))
	}
	BroadcastSubList()
	return
}

//Start sets lobby.State to LobbyStateInProgress, calls SubNotInGamePlayers after 5 minutes
func (lobby *Lobby) Start() {
	db.DB.Table("lobbies").Where("id = ?", lobby.ID).Update("state", LobbyStateInProgress)
	time.AfterFunc(time.Minute*5, func() {
		if lobby.CurrentState() != LobbyStateEnded {
			lobby.SubNotInGamePlayers()
		}
	})
}

// manually called. Should be called after the change to lobby actually takes effect.
func (lobby *Lobby) RealAfterSave() {
	lobby.OnChange(true)
}

//OnChange broadcasts the given lobby to other players. If base is true, broadcasts the lobby list too.
func (lobby *Lobby) OnChange(base bool) {
	switch lobby.State {
	case LobbyStateWaiting, LobbyStateInProgress, LobbyStateReadyingUp:
		BroadcastLobby(lobby)
		if base {
			BroadcastLobbyList()
		}
	}
}

//BroadcastLobby broadcasts the lobby to the lobby's public room (id_public)
func BroadcastLobby(lobby *Lobby) {
	//db.DB.Preload("Spectators").First(&lobby, lobby.ID)
	room := strconv.FormatUint(uint64(lobby.ID), 10)

	broadcaster.SendMessageToRoom(fmt.Sprintf("%s_public", room), "lobbyData", DecorateLobbyData(lobby, true))
}

//BroadcastLobbyToUser broadcasts the lobby to the a user with the given steamID
func BroadcastLobbyToUser(lobby *Lobby, steamid string) {
	//db.DB.Preload("Spectators").First(&lobby, lobby.ID)
	broadcaster.SendMessage(steamid, "lobbyData", DecorateLobbyData(lobby, true))
}

//BroadcastLobbyList broadcasts the lobby list to all users
func BroadcastLobbyList() {
	lobbies := []Lobby{}
	db.DB.Where("state = ?", LobbyStateWaiting).Order("id desc").Find(&lobbies)
	broadcaster.SendMessageToRoom(
		fmt.Sprintf("%s_public", config.GlobalChatRoom),
		"lobbyListData", DecorateLobbyListData(lobbies))
}

func (l *Lobby) LobbyData(include bool) LobbyData {
	return DecorateLobbyData(l, include)
}

//Substitute sets the needs_sub column of the given slot to true, and broadcasts the new
//substitute list
func (lobby *Lobby) Substitute(player *Player) {
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("needs_sub", true)

	db.DB.Preload("Stats").First(player, player.ID)
	player.Stats.IncreaseSubCount()
	BroadcastSubList()
}

//BroadcastSubList broadcasts a the subtitute list to the room 0_public
func BroadcastSubList() {
	subList := DecorateSubstituteList()
	broadcaster.SendMessageToRoom("0_public", "subListData", subList)
}

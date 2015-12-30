// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/jinzhu/gorm"
)

type LobbyType int
type LobbyState int

const (
	LobbyTypeSixes      LobbyType = 6
	LobbyTypeHighlander LobbyType = 9
	LobbyTypeFours      LobbyType = 4
	LobbyTypeUltiduo    LobbyType = 3
	LobbyTypeBball      LobbyType = 2
	LobbyTypeDebug      LobbyType = 1

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

	ReqHoursErr       = helpers.NewTPError("You don't have sufficient hours to join this lobby", 3)
	ReqLobbiesErr     = helpers.NewTPError("You haven't played sufficient lobbies to join this lobby", 3)
	ReqReliabilityErr = helpers.NewTPError("You have insufficient reliability to join this lobby", 3)
)

// Represents an occupied player slot in a lobby
type LobbySlot struct {
	ID       uint // ID of the lobby
	LobbyID  uint // ID of the player occupying the slot
	PlayerID uint // Slot number
	Slot     int  // Denotes if the player is ready
	Ready    bool // Denotes if the player is in game
	InGame   bool // true if the player is in the game server
}

type ServerRecord struct {
	ID             uint
	Host           string
	ServerPassword string // sv_password
	RconPassword   string // rcon_password
}

//Given Lobby IDs are unique, we'll use them for mumble channel names
//
// Represents a Lobby
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

	// TF2 Server Info
	ServerInfo   ServerRecord
	ServerInfoID uint

	Whitelist int //whitelist.tf ID

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

	Hours       int     // minimum hours needed
	Lobbies     int     // minimum lobbies played
	Reliability float64 // minimum reliability needed
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

// GetGlobalRequirement returns the global requirement for the lobby l
func (lobby *Lobby) GetGeneralRequirement() (*Requirement, error) {
	requirement := &Requirement{}
	err := db.DB.Table("requirements").Where("lobby_id = ? AND slot = ?", lobby.ID, -1).First(requirement).Error

	return requirement, err
}

func (lobby *Lobby) GetSlotRequirement(slot int) (*Requirement, error) {
	req := &Requirement{}
	err := db.DB.Table("requirements").Where("lobby_id = ? AND slot = ?", lobby.ID, slot).First(req).Error

	return req, err
}

func (lobby *Lobby) HasSlotRequirement(slot int) bool {
	var count int
	db.DB.Table("requirements").Where("lobby_id = ? AND slot = ?", lobby.ID, slot).Count(&count)
	return count != 0
}

func (lobby *Lobby) HasGeneralRequirement() bool {
	var count int
	db.DB.Table("requirements").Where("lobby_id = ? AND slot = -1", lobby.ID).Count(&count)
	return count != 0
}

// HasRequirements returns true if the given slot has a requirement
func (lobby *Lobby) HasRequirements(slot int) bool {
	return lobby.HasSlotRequirement(slot) || lobby.HasGeneralRequirement()
}

func (l *Lobby) FitsRequirements(player *Player, slot int) (bool, *helpers.TPError) {
	requirements := []*Requirement{}

	player.UpdatePlayerInfo()
	player.Save()

	global, err := l.GetGeneralRequirement()
	if err == nil {
		requirements = append(requirements, global)
	}

	slotReq, err := l.GetSlotRequirement(slot)
	if err == nil {
		requirements = append(requirements, slotReq)
	}

	stats := PlayerStats{}
	db.DB.First(&stats, player.StatsID)

	for _, req := range requirements {
		if player.GameHours < req.Hours {
			return false, ReqHoursErr
		}

		if stats.TotalLobbies() < req.Lobbies {
			return false, ReqLobbiesErr
		}

		//BUG(vibhavp): FitsRequirements doesn't check reliability
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
func NewLobby(mapName string, lobbyType LobbyType, league string, serverInfo ServerRecord, whitelist int, mumble bool, whitelistGroup, password string) *Lobby {
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

// Given a player, returns the slot object if the player occupies a slot in the lobby
func (lobby *Lobby) GetPlayerSlotObj(player *Player) (*LobbySlot, error) {
	slotObj := &LobbySlot{}

	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).First(slotObj).Error

	return slotObj, err
}

// Given a player, returns the slot number if the player occupies a slot int eh lobby
func (lobby *Lobby) GetPlayerSlot(player *Player) (int, error) {
	slotObj, err := lobby.GetPlayerSlotObj(player)

	return slotObj.Slot, err
}

// Returns the ID of the player occupying the slot number
func (lobby *Lobby) GetPlayerIDBySlot(slot int) (uint, error) {
	slotObj := &LobbySlot{}

	err := db.DB.Where("lobby_id = ? AND slot = ?", lobby.ID, slot).First(slotObj).Error

	return uint(slotObj.PlayerID), err
}

// Save changes made to lobby object
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

// Get the lobby object, plus the ServerInfo object inside it
func GetLobbyByIdServer(id uint) (*Lobby, *helpers.TPError) {
	nonExistentLobby := helpers.NewTPError("Lobby not in the database", -1)

	lob := &Lobby{}
	err := db.DB.Preload("ServerInfo").First(lob, id).Error

	if err != nil {
		return nil, nonExistentLobby
	}

	return lob, nil
}

// Get the lobby object, without the ServerInfo object inside it.
func GetLobbyByID(id uint) (*Lobby, *helpers.TPError) {
	nonExistentLobby := helpers.NewTPError("Lobby not in the database", -1)

	lob := &Lobby{}
	err := db.DB.First(lob, id).Error

	if err != nil {
		return nil, nonExistentLobby
	}

	return lob, nil
}

// Add player to lobby, If the player occupies a slot in the lobby already, switch slots.
// If the player is in another lobby, remove them from that lobby before adding them.
func (lobby *Lobby) AddPlayer(player *Player, slot int, password string) *helpers.TPError {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */

	if lobby.SlotPassword != "" && lobby.SlotPassword != password {
		if lobby.PlayerWhitelist == "" && password != "" {
			return InvalidPasswordErr
		}
	}

	if player.ID == 0 {
		return helpers.NewTPError("Player not in the database", -1)
	}

	num := 0

	// It should really be possible to do this query using relations
	if err := db.DB.Table("banned_players_lobbies").
		Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).
		Count(&num).Error; num > 0 || err != nil {
		//helpers.Logger.Debug(fmt.Sprint(err))
		return LobbyBanErr
	}

	if slot >= 2*NumberOfClassesMap[lobby.Type] || slot < 0 {
		return BadSlotErr
	}

	var slotChange bool
	if currLobbyId, err := player.GetLobbyID(false); err == nil {

		if currLobbyId != lobby.ID {
			// if the player is in a different lobby, remove them from that lobby
			curLobby, _ := GetLobbyByID(currLobbyId)

			if curLobby.State == LobbyStateInProgress {
				sub, _ := NewSub(curLobby.ID, player.ID)
				sub.Save()
				BroadcastSubList()
			}

			curLobby.RemovePlayer(player)

		} else {
			// assign the player to a new slot
			// try to remove them from the old slot (in case they are switching slots)
			db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).Delete(&LobbySlot{})
			DisallowPlayer(lobby.ID, player.SteamID)
			slotChange = true
		}
	}

	if !slotChange {
		url := fmt.Sprintf(`http://steamcommunity.com/groups/%s/memberslistxml/?xml=1`,
			lobby.PlayerWhitelist)

		if lobby.PlayerWhitelist != "" && !helpers.IsWhitelisted(player.SteamID, url) {
			return NotWhitelistedErr
		}

		if lobby.HasRequirements(slot) {
			if ok, err := lobby.FitsRequirements(player, slot); !ok {
				return err
			}
		}
	}

	// Check if player is a substitute
	var count int
	db.DB.Table("substitutes").Where("lobby_id = ? AND slot = ? AND filled = ?", lobby.ID, slot, false).Count(&count)
	if count != 0 {
		db.DB.Table("substitutes").Where("lobby_id = ? AND slot = ? AND filled = ?", lobby.ID, slot, false).UpdateColumn("filled", true)
		class, team, _ := LobbyGetSlotInfoString(lobby.Type, slot)
		Say(lobby.ID, fmt.Sprintf("Substitute found for %s %s: %s (%s)", team, class, player.Name, player.SteamID))
		FumbleLobbyPlayerJoinedSub(lobby, player, slot)
	} else if _, err := lobby.GetPlayerIDBySlot(slot); err == nil {
		return FilledErr
	} else {
		FumbleLobbyPlayerJoined(lobby, player, slot)
	}

	//try to remove them from spectators
	lobby.RemoveSpectator(player, false)

	newSlotObj := &LobbySlot{
		PlayerID: player.ID,
		LobbyID:  lobby.ID,
		Slot:     slot,
	}

	db.DB.Create(newSlotObj)

	lobby.OnChange(true)
	if count != 0 {
		BroadcastSubList()
	}

	return nil
}

// Remove a player from the lobby
func (lobby *Lobby) RemovePlayer(player *Player) *helpers.TPError {
	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).Delete(&LobbySlot{}).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}

	DisallowPlayer(lobby.ID, player.SteamID)
	lobby.OnChange(true)
	return nil
}

func (lobby *Lobby) BanPlayer(player *Player) {
	DisallowPlayer(lobby.ID, player.SteamID)
	db.DB.Model(lobby).Association("BannedPlayers").Append(player)
}

// Ready up a player, used when lobby.State == LobbyStateWaiting
func (lobby *Lobby) ReadyPlayer(player *Player) *helpers.TPError {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("ready", true).Error
	if err != nil {
		return helpers.NewTPError("Player is not in the lobby.", 5)
	}
	lobby.OnChange(false)
	return nil
}

// Unreadies a player, used when lobby.State == LobbyStateWaiting
func (lobby *Lobby) UnreadyPlayer(player *Player) *helpers.TPError {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("ready", false).Error
	if err != nil {
		return helpers.NewTPError("Player is not in the lobby.", 5)
	}

	lobby.OnChange(false)
	return nil
}

// Remove players who haven't removed. If spec == true, move them to spectators
func (lobby *Lobby) RemoveUnreadyPlayers(spec bool) error {
	playerids := []uint{}

	if spec {
		err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND ready = ?", lobby.ID, false).Pluck("player_id", &playerids).Error
		if err != nil {
			return err
		}
	}

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

// Returns true if the player is in-game
func (lobby *Lobby) IsPlayerInGame(player *Player) (bool, error) {
	ingame := []bool{}
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).Pluck("in_game", &ingame).Error
	if err != nil {
		return false, err
	}

	return (len(ingame) != 0 && ingame[0]), err
}

// Return true if the player is ready
func (lobby *Lobby) IsPlayerReady(player *Player) (bool, *helpers.TPError) {
	ready := []bool{}
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).Pluck("ready", &ready).Error
	if err != nil {
		return false, helpers.NewTPError("Player is not in the lobby.", 5)
	}
	return (len(ready) != 0 && ready[0]), nil
}

// Unreadies all players
func (lobby *Lobby) UnreadyAllPlayers() error {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).UpdateColumn("ready", false).Error

	lobby.OnChange(false)
	return err
}

// Returns the amount of time left to ready up before a player is kicked.
// Used when a player reconnects while lobby.State == LobbyStateReadyingUp and the player isn't ready
func (lobby *Lobby) ReadyUpTimeLeft() int64 {
	return int64(lobby.ReadyUpTimestamp - time.Now().Unix())
}

// Returns true if all players have readied up.
func (lobby *Lobby) IsEveryoneReady() bool {
	readyPlayers := 0
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND ready = ?", lobby.ID, true).Count(&readyPlayers)

	return readyPlayers == 2*NumberOfClassesMap[lobby.Type]
}

// Adds a player as a lobby spectator
func (lobby *Lobby) AddSpectator(player *Player) *helpers.TPError {
	err := db.DB.Model(lobby).Association("Spectators").Append(player).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}
	lobby.OnChange(false)
	return nil
}

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

// Get number of occupeid slots in the lobby
func (lobby *Lobby) GetPlayerNumber() int {
	count := 0
	err := db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

// Returns true if all lobby spots have been filled
func (lobby *Lobby) IsFull() bool {
	return lobby.GetPlayerNumber() == 2*NumberOfClassesMap[lobby.Type]
}

// Returns true if the given slot (by number) is occupied by a player
func (lobby *Lobby) IsSlotFilled(slot int) bool {
	_, err := lobby.GetPlayerIDBySlot(slot)
	if err != nil {
		return false
	}
	return true
}

// Setup the TF2 server for the lobby
func (lobby *Lobby) SetupServer() error {
	if lobby.State == LobbyStateEnded {
		return nil
	}

	err := SetupServer(lobby.ID, lobby.ServerInfo, lobby.Type, lobby.League, lobby.Whitelist, lobby.MapName)
	return err
}

// Close the lobby.
// Sets lobby.Sate to LobbyStateClosed
// Marks all unfilled substitutes for the lobby as "filled" (so they don't appear in the frontend)
// Delete the TF2 Server Info
// Broadcast the change to all users
func (lobby *Lobby) Close(rpc bool) {
	db.DB.Table("substitutes").Where("lobby_id = ?", lobby.ID).UpdateColumn("filled", true)
	db.DB.First(&lobby).UpdateColumn("state", LobbyStateEnded)
	db.DB.Delete(&lobby.ServerInfo)
	db.DB.Table("requirements").Where("lobby_id = ?", lobby.ID).Delete(&Requirement{})
	//db.DB.Exec("DELETE FROM spectators_players_lobbies WHERE lobby_id = ?", lobby.ID)
	if rpc {
		End(lobby.ID)
	}

	privateRoom := fmt.Sprintf("%d_private", lobby.ID)
	broadcaster.SendMessageToRoom(privateRoom, "lobbyLeft", DecorateLobbyLeave(lobby))

	publicRoom := fmt.Sprintf("%d_public", lobby.ID)
	broadcaster.SendMessageToRoom(publicRoom, "lobbyClosed", DecorateLobbyClosed(lobby))

	BroadcastLobby(lobby)
}

// Update stats (lobbies played count) for
func (lobby *Lobby) UpdateStats() {
	slots := [](*LobbySlot){}
	db.DB.Where("lobby_id = ?", lobby.ID).Find(&slots)

	for _, slot := range slots {
		player := &Player{}
		err := db.DB.First(player, slot.PlayerID).Error
		if err != nil {
			helpers.Logger.Critical("%s", err.Error())
			return
		}
		db.DB.Preload("Stats").First(player, slot.PlayerID)
		player.Stats.PlayedCountIncrease(lobby.Type)
		player.Save()
	}
	lobby.OnChange(false)
}

func (lobby *Lobby) setInGameStatus(player *Player, inGame bool) error {
	err := db.DB.Table("lobby_slots").Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).UpdateColumn("in_game", inGame).Error

	lobby.OnChange(false)
	return err
}

func (lobby *Lobby) SetInGame(player *Player) error {
	return lobby.setInGameStatus(player, true)
}

func (lobby *Lobby) SetNotInGame(player *Player) error {
	return lobby.setInGameStatus(player, false)
}

// Substitute and Remove players not in the game
func (lobby *Lobby) SubNotInGamePlayers() {
	playerids := []uint{}
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND in_game = ?", lobby.ID, false).Pluck("player_id", &playerids).Error
	if err != nil {
		helpers.Logger.Error(err.Error())
		return
	}

	for _, id := range playerids {
		sub, err := NewSub(lobby.ID, id)
		if err != nil {
			helpers.Logger.Error(err.Error())
			continue
		}
		sub.Save()
	}
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND in_game = ?", lobby.ID, false).Delete(&LobbySlot{})
	lobby.OnChange(true)
	BroadcastSubList()
	return
}

// Set lobby.State to LobbyStateInProgress, remove and sub players not-in game after 2 minutes
func (lobby *Lobby) Start() {
	db.DB.Table("lobbies").Where("id = ?", lobby.ID).Update("state", LobbyStateInProgress)
	time.AfterFunc(time.Minute*2, lobby.SubNotInGamePlayers)
}

// manually called. Should be called after the change to lobby actually takes effect.
func (lobby *Lobby) RealAfterSave() {
	lobby.OnChange(true)
}

// If base is true, broadcasts the lobby list update
func (lobby *Lobby) OnChange(base bool) {
	switch lobby.State {
	case LobbyStateWaiting, LobbyStateInProgress, LobbyStateReadyingUp:
		BroadcastLobby(lobby)
		if base {
			BroadcastLobbyList()
		}
	}
}

// Broadcasts the lobby to the lobby public room
func BroadcastLobby(lobby *Lobby) {
	//db.DB.Preload("Spectators").First(&lobby, lobby.ID)
	room := strconv.FormatUint(uint64(lobby.ID), 10)

	broadcaster.SendMessageToRoom(fmt.Sprintf("%s_public", room), "lobbyData", DecorateLobbyData(lobby, true))
}

// Broadcastst the lobby to the user
func BroadcastLobbyToUser(lobby *Lobby, steamid string) {
	//db.DB.Preload("Spectators").First(&lobby, lobby.ID)
	broadcaster.SendMessage(steamid, "lobbyData", DecorateLobbyData(lobby, true))
}

// Broadcasts the lobby list to all users
func BroadcastLobbyList() {
	lobbies := []Lobby{}
	db.DB.Where("state = ?", LobbyStateWaiting).Order("id desc").Find(&lobbies)
	broadcaster.SendMessageToRoom(
		fmt.Sprintf("%s_public", config.Constants.GlobalChatRoom),
		"lobbyListData", DecorateLobbyListData(lobbies))
}

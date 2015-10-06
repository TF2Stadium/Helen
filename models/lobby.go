// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"fmt"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/jinzhu/gorm"
	"strconv"
)

type LobbyType int
type Whitelist int
type LobbyState int

const (
	LobbyTypeSixes      LobbyType = 0
	LobbyTypeHighlander LobbyType = 1
)

var TypePlayerCount = map[LobbyType]int{
	LobbyTypeSixes:      6,
	LobbyTypeHighlander: 9,
}

const (
	LobbyStateInitializing LobbyState = 0
	LobbyStateWaiting      LobbyState = 1
	LobbyStateReadyingUp   LobbyState = 2
	LobbyStateInProgress   LobbyState = 3
	LobbyStateEnded        LobbyState = 5
)

var LobbyServerSettingUp = make(map[uint]time.Time)

var stateString = map[LobbyState]string{
	LobbyStateWaiting:    "Waiting For Players",
	LobbyStateInProgress: "Lobby in Progress",
	LobbyStateEnded:      "Lobby Ended",
}

var FormatMap = map[LobbyType]string{
	LobbyTypeSixes:      "Sixes",
	LobbyTypeHighlander: "Highlander",
}

var ValidLeagues = map[string]bool{
	"ugc":   true,
	"etf2l": true,
}

var readyUpLobbyID = make(chan uint)

type LobbySlot struct {
	ID uint
	// Lobby    Lobby
	LobbyId uint
	// Player   Player
	PlayerId uint
	Slot     int
	Ready    bool
	InGame   bool
}

type ServerRecord struct {
	ID             uint
	Host           string
	ServerPassword string
	RconPassword   string
}

//Given Lobby IDs are unique, we'll use them for mumble channel names
type Lobby struct {
	gorm.Model
	MapName string
	State   LobbyState
	Type    LobbyType
	League  string

	Slots []LobbySlot

	ServerInfo   ServerRecord
	ServerInfoID uint

	Whitelist Whitelist //whitelist.tf ID

	Spectators []Player `gorm:"many2many:spectators_players_lobbies"`

	BannedPlayers []Player `gorm:"many2many:banned_players_lobbies"`

	CreatedByID uint
	CreatedBy   Player
}

func NewLobby(mapName string, lobbyType LobbyType, league string, serverInfo ServerRecord, whitelist int) *Lobby {
	lobby := &Lobby{
		Type:       lobbyType,
		State:      LobbyStateInitializing,
		League:     league,
		MapName:    mapName,
		Whitelist:  Whitelist(whitelist), // that's a strange line
		ServerInfo: serverInfo,
	}

	// Must specify CreatedBy manually if the lobby is created by a player

	return lobby
}

func (lobby *Lobby) GetPlayerSlot(player *Player) (int, error) {
	slotObj := &LobbySlot{}

	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).First(slotObj).Error

	return slotObj.Slot, err
}

func (lobby *Lobby) GetPlayerIdBySlot(slot int) (uint, error) {
	slotObj := &LobbySlot{}

	err := db.DB.Where("lobby_id = ? AND slot = ?", lobby.ID, slot).First(slotObj).Error

	return uint(slotObj.PlayerId), err
}

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

func GetLobbyById(id uint) (*Lobby, *helpers.TPError) {
	nonExistentLobby := helpers.NewTPError("Lobby not in the database", -1)

	lob := &Lobby{}
	err := db.DB.Preload("ServerInfo").First(lob, id).Error

	if err != nil {
		return nil, nonExistentLobby
	}

	return lob, nil
}

// //Add player to lobby
func (lobby *Lobby) AddPlayer(player *Player, slot int) *helpers.TPError {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */

	lobbyBanError := helpers.NewTPError("The player has been banned from this lobby.", 4)
	badSlotError := helpers.NewTPError("This slot does not exist.", 3)
	filledError := helpers.NewTPError("This slot has been filled.", 2)
	alreadyInLobbyError := helpers.NewTPError("Player is already in a lobby", 1)

	if player.ID == 0 {
		return helpers.NewTPError("Player not in the database", -1)
	}

	num := 0

	// It should really be possible to do this query using relations
	if err := db.DB.Table("banned_players_lobbies").
		Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).
		Count(&num).Error; num > 0 || err != nil {
		helpers.Logger.Debug(fmt.Sprint(err))
		return lobbyBanError
	}

	if slot >= 2*TypePlayerCount[lobby.Type] || slot < 0 {
		return badSlotError
	}

	slotFilled := false
	if _, err := lobby.GetPlayerIdBySlot(slot); err == nil {
		slotFilled = true
	}

	currLobbyId, err := player.GetLobbyId()

	// if the player is in a different lobby, return error
	if err == nil && currLobbyId != lobby.ID {
		return alreadyInLobbyError
	}

	// if the slot is occupied, return error
	if slotFilled {
		return filledError
	}

	// assign the player to a new slot
	// try to remove them from the old slot (in case they are switching slots)
	lobby.RemovePlayer(player)
	// try to remove them from spectators
	lobby.RemoveSpectator(player)

	newSlotObj := &LobbySlot{
		PlayerId: player.ID,
		LobbyId:  lobby.ID,
		Slot:     slot,
	}

	db.DB.Create(newSlotObj)

	AllowPlayer(lobby.ID, player.SteamId)
	lobby.OnChange(true)
	return nil
}

func (lobby *Lobby) RemovePlayer(player *Player) *helpers.TPError {
	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).Delete(&LobbySlot{}).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}

	lobby.OnChange(true)
	return nil
}

func (lobby *Lobby) BanPlayer(player *Player) {
	DisallowPlayer(lobby.ID, player.SteamId)
	db.DB.Model(lobby).Association("BannedPlayers").Append(player)
}

func (lobby *Lobby) ReadyPlayer(player *Player) *helpers.TPError {
	slot := &LobbySlot{}
	err := db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(slot).Error
	if err != nil {
		return helpers.NewTPError("Player is not in the lobby.", 5)
	}
	slot.Ready = true
	db.DB.Save(slot)
	lobby.OnChange(false)
	return nil
}

func (lobby *Lobby) UnreadyPlayer(player *Player) *helpers.TPError {
	slot := &LobbySlot{}
	err := db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(slot).Error
	if err != nil {
		return helpers.NewTPError("Player is not in the lobby.", 5)
	}

	slot.Ready = false
	db.DB.Save(slot)
	lobby.OnChange(false)
	return nil
}

func (lobby *Lobby) RemoveUnreadyPlayers() error {
	err := db.DB.Where("lobby_id = ? AND ready = ?", lobby.ID, false).Delete(&LobbySlot{}).Error
	lobby.OnChange(true)
	return err
}

func (lobby *Lobby) IsPlayerReady(player *Player) (bool, *helpers.TPError) {
	slot := &LobbySlot{}
	err := db.DB.Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).First(slot).Error
	if err != nil {
		return false, helpers.NewTPError("Player is not in the lobby.", 5)
	}
	return slot.Ready, nil
}

func (lobby *Lobby) UnreadyAllPlayers() error {
	var slots []LobbySlot
	err := db.DB.Where("lobby_id = ?", lobby.ID).Find(&slots).Error
	for _, slot := range slots {
		slot.Ready = false
		db.DB.Save(slot)
	}
	lobby.OnChange(false)
	return err
}

func ReadyTimeoutListener() {
	for {
		select {
		case id := <-readyUpLobbyID:
			tick := time.After(time.Second * 30)
			<-tick

			lobby := &Lobby{}
			db.DB.First(lobby, id)

			if lobby.State != LobbyStateInProgress {
				helpers.LockRecord(lobby.ID, lobby)
				defer helpers.UnlockRecord(lobby.ID, lobby)
				err := lobby.RemoveUnreadyPlayers()
				if err != nil {
					helpers.Logger.Critical(err.Error())
				}

				lobby.UnreadyAllPlayers()
				if err != nil {
					helpers.Logger.Critical(err.Error())
				}

				lobby.State = LobbyStateWaiting
				lobby.Save()
			}
		}
	}
}

func (lobby *Lobby) ReadyUpTimeoutCheck() {
	readyUpLobbyID <- lobby.ID
}

func (lobby *Lobby) IsEveryoneReady() bool {
	var slots []LobbySlot
	db.DB.Where("lobby_id = ?", lobby.ID).Find(&slots)

	if len(slots) != TypePlayerCount[lobby.Type]*2 {
		return false
	}

	for _, slot := range slots {
		if !slot.Ready {
			return false
		}
	}
	return true
}

func (lobby *Lobby) AddSpectator(player *Player) *helpers.TPError {
	if _, err := lobby.GetPlayerSlot(player); err == nil {
		return lobby.RemovePlayer(player)
	}

	err := db.DB.Model(lobby).Association("Spectators").Append(player).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}
	lobby.OnChange(false)
	return nil
}

func (lobby *Lobby) RemoveSpectator(player *Player) *helpers.TPError {
	err := db.DB.Model(lobby).Association("Spectators").Delete(player).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}
	lobby.OnChange(false)
	return nil
}

func (lobby *Lobby) GetPlayerNumber() int {
	count := 0
	err := db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

func (lobby *Lobby) IsFull() bool {
	return lobby.GetPlayerNumber() >= 2*TypePlayerCount[lobby.Type]
}

func (lobby *Lobby) IsSlotFilled(slot int) bool {
	_, err := lobby.GetPlayerIdBySlot(slot)
	if err != nil {
		return false
	}
	return true
}

func (lobby *Lobby) SetupServer() error {
	if lobby.State == LobbyStateEnded {
		return nil
	}

	err := SetupServer(lobby.ID, lobby.ServerInfo, lobby.Type, lobby.League, lobby.MapName)
	if err != nil {
		return helpers.NewTPError(err.Error(), 0)
	}
	return nil
}

func (lobby *Lobby) Close(rpc bool) {
	lobby.State = LobbyStateEnded
	db.DB.Delete(&lobby.ServerInfo)
	if rpc {
		End(lobby.ID)
	}
	delete(LobbyServerSettingUp, lobby.ID)
	db.DB.Save(lobby)
	helpers.RemoveRecord(lobby.ID, lobby)
}

// GORM callback
func (lobby *Lobby) AfterFind() error {
	if (lobby.ServerInfo == ServerRecord{}) {
		// hasn't been preloaded. Do that here.
		db.DB.Find(&lobby.ServerInfo, lobby.ServerInfoID)
	}

	return nil
}

// manually called. Should be called after the change to lobby actually takes effect.
func (lobby *Lobby) RealAfterSave() {
	lobby.OnChange(true)
}

// If base is true, broadcasts the lobby list update
func (lobby *Lobby) OnChange(base bool) {
	if lobby.State == LobbyStateWaiting || lobby.State == LobbyStateInProgress ||
		lobby.State == LobbyStateReadyingUp {
		BroadcastLobby(lobby)
	}

	if lobby.State == LobbyStateWaiting && base {
		BroadcastLobbyList()
	}
}

func IsLeagueValid(league string) bool {
	_, ok := ValidLeagues[league]
	return ok
}

func BroadcastLobby(lobby *Lobby) {
	bytes, _ := DecorateLobbyDataJSON(lobby).Encode()
	broadcaster.SendMessageToRoom(strconv.FormatUint(uint64(lobby.ID), 10), "lobbyData", string(bytes))
}

func BroadcastLobbyToUser(lobby *Lobby, steamid string) {
	bytes, _ := DecorateLobbyDataJSON(lobby).Encode()
	broadcaster.SendMessage(steamid, "lobbyData", string(bytes))
}

func BroadcastLobbyList() {
	var lobbies []Lobby
	db.DB.Where("state = ?", LobbyStateWaiting).Order("id desc").Find(&lobbies)
	list, err := DecorateLobbyListData(lobbies)
	if err != nil {
		helpers.Logger.Warning("Failed to send lobby list: %s", err.Error())
	} else {
		broadcaster.SendMessageToRoom(config.Constants.GlobalChatRoom, "lobbyListData", list)
	}
}

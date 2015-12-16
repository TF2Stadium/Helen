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
)

const (
	LobbyStateInitializing LobbyState = 0
	LobbyStateWaiting      LobbyState = 1
	LobbyStateReadyingUp   LobbyState = 2
	LobbyStateInProgress   LobbyState = 3
	LobbyStateEnded        LobbyState = 5
)

var stateString = map[LobbyState]string{
	LobbyStateWaiting:    "Waiting For Players",
	LobbyStateInProgress: "Lobby in Progress",
	LobbyStateEnded:      "Lobby Ended",
}

var FormatMap = map[LobbyType]string{
	LobbyTypeSixes:      "6s",
	LobbyTypeHighlander: "Highlander",
	LobbyTypeFours:      "4v4",
	LobbyTypeUltiduo:    "Ultiduo",
	LobbyTypeBball:      "Bball",
	LobbyTypeDebug:      "Debug",
}

type LobbySlot struct {
	ID uint
	// Lobby    Lobby
	LobbyId uint
	// Player   Player
	PlayerId uint
	Slot     int
	Ready    bool
	InGame   bool

	Team  string
	Class string
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
	Mode    string
	MapName string
	State   LobbyState
	Type    LobbyType
	League  string

	RegionCode string
	RegionName string

	Mumble bool

	Slots []LobbySlot

	ServerInfo   ServerRecord
	ServerInfoID uint

	Whitelist int //whitelist.tf ID

	Spectators []Player `gorm:"many2many:spectators_players_lobbies"`

	BannedPlayers []Player `gorm:"many2many:banned_players_lobbies"`

	CreatedBySteamID string

	ReadyUpTimestamp int64 //Stores the timestamp at which the ready up timeout started
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

func NewLobby(mapName string, lobbyType LobbyType, league string, serverInfo ServerRecord, whitelist int, mumble bool) *Lobby {
	lobby := &Lobby{
		Mode:       getGamemode(mapName, lobbyType),
		Type:       lobbyType,
		State:      LobbyStateInitializing,
		League:     league,
		MapName:    mapName,
		Whitelist:  whitelist, // that's a strange line
		Mumble:     mumble,
		ServerInfo: serverInfo,
	}

	// Must specify CreatedBy manually if the lobby is created by a player

	return lobby
}

func (lobby *Lobby) GetPlayerSlotObj(player *Player) (*LobbySlot, error) {
	slotObj := &LobbySlot{}

	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).First(slotObj).Error

	return slotObj, err
}

func (lobby *Lobby) GetPlayerSlot(player *Player) (int, error) {
	slotObj, err := lobby.GetPlayerSlotObj(player)

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

func GetLobbyByIdServer(id uint) (*Lobby, *helpers.TPError) {
	nonExistentLobby := helpers.NewTPError("Lobby not in the database", -1)

	lob := &Lobby{}
	err := db.DB.Preload("ServerInfo").First(lob, id).Error

	if err != nil {
		return nil, nonExistentLobby
	}

	return lob, nil
}

func GetLobbyById(id uint) (*Lobby, *helpers.TPError) {
	nonExistentLobby := helpers.NewTPError("Lobby not in the database", -1)

	lob := &Lobby{}
	err := db.DB.First(lob, id).Error

	if err != nil {
		return nil, nonExistentLobby
	}

	return lob, nil
}

// //Add player to lobby
func (lobby *Lobby) AddPlayer(player *Player, slot int, team, class string) *helpers.TPError {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */

	lobbyBanError := helpers.NewTPError("The player has been banned from this lobby.", 4)
	badSlotError := helpers.NewTPError("This slot does not exist.", 3)
	filledError := helpers.NewTPError("This slot has been filled.", 2)

	if player.ID == 0 {
		return helpers.NewTPError("Player not in the database", -1)
	}

	num := 0

	// It should really be possible to do this query using relations
	if err := db.DB.Table("banned_players_lobbies").
		Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).
		Count(&num).Error; num > 0 || err != nil {
		//helpers.Logger.Debug(fmt.Sprint(err))
		return lobbyBanError
	}

	if slot >= 2*NumberOfClassesMap[lobby.Type] || slot < 0 {
		return badSlotError
	}

	if currLobbyId, err := player.GetLobbyId(); err == nil {
		if currLobbyId != lobby.ID {
			// if the player is in a different lobby, remove them from that lobby
			curLobby, _ := GetLobbyById(currLobbyId)
			if curLobby.State == LobbyStateInProgress {
				sub, _ := NewSub(curLobby.ID, player.SteamId)
				db.DB.Save(sub)
				BroadcastSubList()
			}
			curLobby.RemovePlayer(player)
		} else {
			// assign the player to a new slot
			// try to remove them from the old slot (in case they are switching slots)
			db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).Delete(&LobbySlot{})
			DisallowPlayer(lobby.ID, player.SteamId)
		}
	}

	var count int
	db.DB.Table("substitutes").Where("lobby_id = ? AND team = ? AND class = ? AND filled = ?", lobby.ID, team, class, false).Count(&count)
	if count != 0 {
		db.DB.Table("substitutes").Where("lobby_id = ? AND team = ? AND class = ? AND filled = ?", lobby.ID, team, class, false).UpdateColumn("filled", true)
		FumbleLobbyPlayerJoinedSub(lobby, player, slot)
	} else if _, err := lobby.GetPlayerIdBySlot(slot); err == nil {
		return filledError
	} else {
		FumbleLobbyPlayerJoined(lobby, player, slot)
	}

	//try to remove them from spectators
	lobby.RemoveSpectator(player, false)

	newSlotObj := &LobbySlot{
		PlayerId: player.ID,
		LobbyId:  lobby.ID,
		Slot:     slot,
		Team:     team,
		Class:    class,
	}

	db.DB.Create(newSlotObj)

	lobby.OnChange(true)
	if count != 0 {
		BroadcastSubList()
	}

	return nil
}

func (lobby *Lobby) RemovePlayer(player *Player) *helpers.TPError {
	err := db.DB.Where("player_id = ? AND lobby_id = ?", player.ID, lobby.ID).Delete(&LobbySlot{}).Error
	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}

	DisallowPlayer(lobby.ID, player.SteamId)
	lobby.OnChange(true)
	return nil
}

func (lobby *Lobby) BanPlayer(player *Player) {
	DisallowPlayer(lobby.ID, player.SteamId)
	db.DB.Model(lobby).Association("BannedPlayers").Append(player)
}

func (lobby *Lobby) ReadyPlayer(player *Player) *helpers.TPError {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("ready", true).Error
	if err != nil {
		return helpers.NewTPError("Player is not in the lobby.", 5)
	}
	lobby.OnChange(false)
	return nil
}

func (lobby *Lobby) UnreadyPlayer(player *Player) *helpers.TPError {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).UpdateColumn("ready", false).Error
	if err != nil {
		return helpers.NewTPError("Player is not in the lobby.", 5)
	}

	lobby.OnChange(false)
	return nil
}

func (lobby *Lobby) RemoveUnreadyPlayers() error {
	err := db.DB.Where("lobby_id = ? AND ready = ?", lobby.ID, false).Delete(&LobbySlot{}).Error
	lobby.OnChange(true)
	return err
}

func (lobby *Lobby) IsPlayerInGame(player *Player) (bool, error) {
	var ingame []bool
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).Pluck("in_game", &ingame).Error
	if err != nil {
		return false, err
	}

	return (len(ingame) != 0 && ingame[0]), err
}

func (lobby *Lobby) IsPlayerReady(player *Player) (bool, *helpers.TPError) {
	var ready []bool
	err := db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", lobby.ID, player.ID).Pluck("ready", &ready).Error
	if err != nil {
		return false, helpers.NewTPError("Player is not in the lobby.", 5)
	}
	return (len(ready) != 0 && ready[0]), nil
}

func (lobby *Lobby) UnreadyAllPlayers() error {
	err := db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).UpdateColumn("ready", false).Error

	lobby.OnChange(false)
	return err
}

func (lobby *Lobby) ReadyUpTimeLeft() int64 {
	return int64(lobby.ReadyUpTimestamp - time.Now().Unix())
}

func (lobby *Lobby) IsEveryoneReady() bool {
	readyPlayers := 0
	db.DB.Table("lobby_slots").Where("lobby_id = ? AND ready = ?", lobby.ID, true).Count(&readyPlayers)

	return readyPlayers == 2*NumberOfClassesMap[lobby.Type]
}

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

func (lobby *Lobby) GetPlayerNumber() int {
	count := 0
	err := db.DB.Table("lobby_slots").Where("lobby_id = ?", lobby.ID).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

func (lobby *Lobby) IsFull() bool {
	return lobby.GetPlayerNumber() >= 2*NumberOfClassesMap[lobby.Type]
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

	err := SetupServer(lobby.ID, lobby.ServerInfo, lobby.Type, lobby.League, lobby.Whitelist, lobby.MapName)
	return err
}

func (lobby *Lobby) Close(rpc bool) {
	db.DB.Table("substitutes").Where("lobby_id = ?", lobby.ID).UpdateColumn("filled", true)
	db.DB.First(&lobby).UpdateColumn("state", LobbyStateEnded)
	db.DB.Delete(&lobby.ServerInfo)
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

func (lobby *Lobby) UpdateStats() {
	var slots []LobbySlot
	db.DB.Where("lobby_id = ?", lobby.ID).Find(&slots)

	for _, slot := range slots {
		player := &Player{}
		err := db.DB.First(player, slot.PlayerId).Error
		if err != nil {
			helpers.Logger.Critical("%s", err.Error())
			return
		}
		db.DB.Preload("Stats").First(player, slot.PlayerId)
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

func BroadcastLobby(lobby *Lobby) {
	//db.DB.Preload("Spectators").First(&lobby, lobby.ID)
	room := strconv.FormatUint(uint64(lobby.ID), 10)

	broadcaster.SendMessageToRoom(fmt.Sprintf("%s_public", room), "lobbyData", DecorateLobbyData(lobby, true))
}

func BroadcastLobbyToUser(lobby *Lobby, steamid string) {
	//db.DB.Preload("Spectators").First(&lobby, lobby.ID)
	broadcaster.SendMessage(steamid, "lobbyData", DecorateLobbyData(lobby, true))
}

func BroadcastLobbyList() {
	var lobbies []Lobby
	db.DB.Where("state = ?", LobbyStateWaiting).Order("id desc").Find(&lobbies)
	broadcaster.SendMessageToRoom(
		fmt.Sprintf("%s_public", config.Constants.GlobalChatRoom),
		"lobbyListData", DecorateLobbyListData(lobbies))
}

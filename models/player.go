// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"time"

	"github.com/TF2Stadium/Helen/config"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/PlayerStatsScraper"
	"github.com/jinzhu/gorm"
)

// Stores types of bans
type PlayerBanType int

const (
	// Ban player from joining lobbies
	PlayerBanJoin PlayerBanType = iota
	// Ban player from creating lobbies
	PlayerBanCreate
	// Ban player from sending chat messages
	PlayerBanChat
	// Complete player ban
	PlayerBanFull
)

//PlayerBan represents a player ban
type PlayerBan struct {
	gorm.Model
	PlayerID uint          // ID of the player banned
	Type     PlayerBanType // Ban type
	Until    time.Time     // Time until which the ban is valid
	Reason   string        // Reason for the ban
	Active   bool          `sql:"default:true"` // Whether the ban is active
}

//PlayerSetting represents a stored player setting
type PlayerSetting struct {
	ID        uint
	CreatedAt time.Time
	PlayerID  uint

	Key   string // Setting key
	Value string `sql:"size:65535"` // Setting value
}

//Player represents a player object
type Player struct {
	gorm.Model
	Debug   bool   // true if player is a dummy one.
	SteamID string `sql:"unique"` // Players steam ID
	Stats   PlayerStats
	StatsID uint

	// info from steam api
	Avatar     string
	Profileurl string
	GameHours  int
	Name       string             // Player name
	Role       authority.AuthRole `sql:"default:0"` // Role is player by default

	Settings       []PlayerSetting
	MumbleUsername string `sql:"unique"`
	MumbleAuthkey  string `sql:"not null;unique"`
}

// Create a new player with the given steam id.
// Use (*Player).Save() to save the player object.
func NewPlayer(steamId string) (*Player, error) {
	player := &Player{SteamID: steamId}

	if !config.Constants.SteamApiMockUp {
		player.Stats = NewPlayerStats()

		err := player.UpdatePlayerInfo()
		if err != nil {
			return &Player{}, err
		}
	} else {
		player.Stats = PlayerStats{}
	}

	return player, nil
}

//if the player has an alias, return that. Else, return their steam name
func (p *Player) Alias() string {
	alias, err := p.GetSetting("siteAlias")
	if err != nil {
		return p.Name
	}
	if alias.Value != "" {
		return alias.Value
	}
	return p.Name
}

// Save any changes made to the player object
func (player *Player) Save() error {
	var err error
	if db.DB.NewRecord(player) {
		err = db.DB.Create(player).Error
	} else {
		err = db.DB.Save(player).Error
	}
	return err
}

// Get a player by it's ID
func GetPlayerByID(ID uint) (*Player, error) {
	player := &Player{}

	if err := db.DB.First(player, ID).Error; err != nil {
		return nil, err
	}

	return player, nil
}

// Get a player object by it's Steam id
func GetPlayerBySteamID(steamid string) (*Player, *helpers.TPError) {
	var player = Player{}
	err := db.DB.Where("steam_id = ?", steamid).First(&player).Error
	if err != nil {
		return nil, helpers.NewTPError("Player is not in the database", -1)
	}
	return &player, nil
}

// Get a player object by it's Steam ID, with the Stats field
func GetPlayerWithStats(steamid string) (*Player, *helpers.TPError) {
	var player = Player{}
	err := db.DB.Where("steam_id = ?", steamid).Preload("Stats").First(&player).Error
	if err != nil {
		return nil, helpers.NewTPError("Player is not in the database", -1)
	}
	return &player, nil
}

// Get the ID of the lobby the player occupies a slot in. Only works for lobbies which aren't closed (LobbyStateEnded).
//If inProrgess, exclude lobbies which are in progress
func (player *Player) GetLobbyID(inProgress bool) (uint, *helpers.TPError) {
	playerSlot := &LobbySlot{}
	states := []LobbyState{LobbyStateEnded}
	if inProgress {
		states = append(states, LobbyStateInProgress)
	}

	err := db.DB.Joins("INNER JOIN lobbies ON lobbies.id = lobby_slots.lobby_id").
		Where("lobby_slots.player_id = ? AND lobbies.state NOT IN (?)", player.ID, states).
		Find(playerSlot).Error

	// if the player is not in any lobby, return error
	if err != nil {
		return 0, helpers.NewTPError("Player not in any lobby", 1)
	}

	return playerSlot.LobbyID, nil
}

// Return true if player is spectating a lobby with the given lobby ID
func (player *Player) IsSpectatingID(lobbyid uint) bool {
	count := 0
	err := db.DB.Table("spectators_players_lobbies").Where("player_id = ? AND lobby_id = ?", player.ID, lobbyid).Count(&count).Error
	if err != nil {
		return false
	}
	return count != 0
}

//Get ID(s) of lobbies (which aren't clsoed) the player is spectating
func (player *Player) GetSpectatingIds() ([]uint, *helpers.TPError) {
	var ids []uint
	err := db.DB.Model(&Lobby{}).
		Joins("INNER JOIN spectators_players_lobbies l ON l.lobby_id = lobbies.id").
		Where("l.player_id = ? AND lobbies.state <> ?", player.ID, LobbyStateEnded).
		Pluck("id", &ids).Error

	if err != nil {
		return nil, helpers.NewTPError(err.Error(), 1)
	}

	return ids, nil
}

//Retrieve the player's details using the Steam API. The object needs to be saved after this.
func (player *Player) UpdatePlayerInfo() error {
	if config.Constants.SteamApiMockUp {
		return nil
	}

	scraper.SetSteamApiKey(config.Constants.SteamDevApiKey)
	p, _ := GetPlayerBySteamID(player.SteamID)

	if p != nil {
		*player = *p
	}

	playerInfo, infoErr := scraper.GetPlayerInfo(player.SteamID)
	if infoErr != nil {
		return infoErr
	}

	// profile state is 1 when the player have a steam community profile
	if playerInfo.Profilestate == 1 && playerInfo.Visibility == "public" {
		pHours, hErr := scraper.GetTF2Hours(player.SteamID)

		if hErr != nil {
			return hErr
		}

		player.GameHours = pHours
	}

	player.Profileurl = playerInfo.Profileurl
	player.Avatar = playerInfo.Avatar
	player.Name = playerInfo.Name

	return nil
}

func (player *Player) SetSetting(key string, value string) error {
	setting := PlayerSetting{}
	err := db.DB.Where("player_id = ? AND key = ?", player.ID, key).First(&setting).Error

	setting.PlayerID = player.ID
	setting.Key = key
	setting.Value = value

	err = db.DB.Save(&setting).Error

	return err
}

func (player *Player) GetSetting(key string) (PlayerSetting, error) {
	setting := PlayerSetting{}
	err := db.DB.Where("player_id = ? AND key = ?", player.ID, key).First(&setting).Error

	return setting, err
}

func (player *Player) GetSettings() ([]PlayerSetting, error) {
	var settings []PlayerSetting
	err := db.DB.Where("player_id = ?", player.ID).Find(&settings).Error

	return settings, err
}

func (player *Player) IsBannedWithTime(t PlayerBanType) (bool, time.Time) {
	ban := &PlayerBan{}
	err := db.DB.Where("type = ? AND until > now() AND player_id = ? AND active = TRUE", t, player.ID).
		Order("until desc").First(ban).Error
	if err != nil {
		return false, time.Time{}
	}

	return true, ban.Until
}

func (player *Player) IsBanned(t PlayerBanType) bool {
	res, _ := player.IsBannedWithTime(t)
	return res
}

func (player *Player) BanUntil(tim time.Time, t PlayerBanType, reason string) error {
	ban := PlayerBan{
		PlayerID: player.ID,
		Type:     t,
		Until:    tim,
		Reason:   reason,
	}

	return db.DB.Create(&ban).Error
}

func (player *Player) Unban(t PlayerBanType) error {
	return db.DB.Model(&PlayerBan{}).Where("player_id = ? AND type = ? AND active = TRUE", player.ID, t).
		Update("active", "FALSE").Error
}

func (player *Player) GetActiveBans() ([]*PlayerBan, error) {
	var bans []*PlayerBan
	err := db.DB.Where("player_id = ? AND active = TRUE AND until > now()", player.ID).Find(&bans).Error
	if err != nil {
		return nil, err
	}
	return bans, nil
}

func GetAllActiveBans() []*PlayerBan {
	var bans []*PlayerBan
	db.DB.Where("active = TRUE AND until > now()").Find(&bans)
	return bans
}

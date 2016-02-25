// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
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

var ErrPlayerNotFound = errors.New("Player not found")

//PlayerBan represents a player ban
type PlayerBan struct {
	gorm.Model
	PlayerID uint          // ID of the player banned
	Type     PlayerBanType // Ban type
	Until    time.Time     // Time until which the ban is valid
	Reason   string        // Reason for the ban
	Active   bool          `sql:"default:true"` // Whether the ban is active
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

	Settings gorm.Hstore

	MumbleUsername string `sql:"unique"`
	MumbleAuthkey  string `sql:"not null;unique"`

	TwitchAccessToken string
	TwitchName        string

	ExternalLinks gorm.Hstore
}

// Create a new player with the given steam id.
// Use (*Player).Save() to save the player object.
func NewPlayer(steamId string) (*Player, error) {
	player := &Player{SteamID: steamId}

	if config.Constants.SteamDevAPIKey == "" {
		player.Stats = NewPlayerStats()

		err := player.UpdatePlayerInfo()
		if err != nil {
			return &Player{}, err
		}
	} else {
		player.Stats = PlayerStats{}
	}

	player.MumbleUsername = player.GenMumbleUsername()
	player.MumbleAuthkey = player.GenAuthKey()

	return player, nil
}

func (player *Player) SetExternalLinks() {
	player.ExternalLinks = make(gorm.Hstore)
	defer player.Save()

	// logs.tf
	logstf := fmt.Sprintf(`http://logs.tf/profile/%s`, player.SteamID)
	resp, err := helpers.HTTPClient.Get(logstf)
	if err == nil && resp.StatusCode == 200 {
		player.ExternalLinks["logstf"] = &logstf
	}

	// UGC
	ugc := fmt.Sprintf(`http://www.ugcleague.com/players_page.cfm?player_id=%s`, player.SteamID)
	resp, err = helpers.HTTPClient.Get(ugc)
	if err == nil && resp.StatusCode == 200 {
		player.ExternalLinks["ugc"] = &ugc
	}

	var reply struct {
		Player *struct {
			ID      int    `json:"id"`
			Country string `json:"country"`
		} `json:"player,omitempty"`
		Status struct {
			Code int `json:"code"`
		}
	}

	etf2lURL := fmt.Sprintf(`http://api.etf2l.org/player/%s`, player.SteamID)
	req, _ := http.NewRequest("GET", etf2lURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err = helpers.HTTPClient.Do(req)
	if err != nil {
		logrus.Error(err)
		return
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&reply)
	if err != nil {
		logrus.Error(err)
		return
	}

	if reply.Player != nil {
		url := fmt.Sprintf(`http://beta.etf2l.org/forum/user/%d/`, reply.Player.ID)
		player.ExternalLinks["etf2l"] = &url
	}

	// teamfortress.tv
	tftv := fmt.Sprintf("http://www.teamfortress.tv/api/users/%s", player.SteamID)
	resp, err = helpers.HTTPClient.Get(tftv)
	if err != nil {
		logrus.Error(err)
		return
	}

	var tftvReply struct {
		UserName string `json:"user_name"`
	}

	dec = json.NewDecoder(resp.Body)
	err = dec.Decode(&reply)
	if err != nil {
		logrus.Error(err)
		return
	}

	if tftvReply.UserName != "" {
		uname := fmt.Sprintf("http://teamfortress.tv/user/%s", tftvReply.UserName)
		player.ExternalLinks["tftv"] = &uname
	}

}

func (player *Player) GenAuthKey() string {
	var count int
	var authKey string

	for {
		buff := bytes.NewBufferString(player.SteamID)
		buff.Grow(32)
		rand.Read(buff.Bytes())

		sum := sha256.Sum256(buff.Bytes())
		authKey = hex.EncodeToString(sum[:])

		db.DB.Table("players").Where("mumble_authkey = ?", authKey).Count(&count)
		if count == 0 {
			break
		}
	}

	return authKey
}

func (player *Player) GenMumbleUsername() string {
	last := &Player{}
	db.DB.Table("players").Last(last)

	mumbleNick := fmt.Sprintf("TF2Stadium%d", last.ID+1)
	return mumbleNick
}

//if the player has an alias, return that. Else, return their steam name
func (p *Player) Alias() string {
	alias := p.GetSetting("siteAlias")
	if alias == "" {
		return p.Name
	}

	return alias
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
func GetPlayerBySteamID(steamid string) (*Player, error) {
	var player = Player{}
	err := db.DB.Where("steam_id = ?", steamid).First(&player).Error
	if err != nil {
		return nil, ErrPlayerNotFound
	}
	return &player, nil
}

// Get a player object by it's Steam ID, with the Stats field
func GetPlayerWithStats(steamid string) (*Player, error) {
	var player = Player{}
	err := db.DB.Where("steam_id = ?", steamid).Preload("Stats").First(&player).Error
	if err != nil {
		return nil, ErrPlayerNotFound
	}
	return &player, nil
}

// Get the ID of the lobby the player occupies a slot in. Only works for lobbies which aren't closed (LobbyStateEnded).
//If inProrgess, exclude lobbies which are in progress
func (player *Player) GetLobbyID(inProgress bool) (uint, error) {
	playerSlot := &LobbySlot{}
	states := []LobbyState{LobbyStateEnded}
	if inProgress {
		states = append(states, LobbyStateInProgress)
	}

	err := db.DB.Joins("INNER JOIN lobbies ON lobbies.id = lobby_slots.lobby_id").
		Where("lobby_slots.player_id = ? AND lobbies.state NOT IN (?)", player.ID, states).
		First(playerSlot).Error

	// if the player is not in any lobby or the player has been reported/needs sub, return error
	if err != nil || playerSlot.NeedsSub {
		return 0, ErrPlayerNotFound
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
func (player *Player) GetSpectatingIds() ([]uint, error) {
	var ids []uint
	err := db.DB.Model(&Lobby{}).
		Joins("INNER JOIN spectators_players_lobbies l ON l.lobby_id = lobbies.id").
		Where("l.player_id = ? AND lobbies.state <> ?", player.ID, LobbyStateEnded).
		Pluck("id", &ids).Error

	if err != nil {
		return nil, err
	}

	return ids, nil
}

//Retrieve the player's details using the Steam API. The object needs to be saved after this.
func (player *Player) UpdatePlayerInfo() error {
	if config.Constants.SteamDevAPIKey == "" {
		return nil
	}

	defer player.Save()
	player.SetExternalLinks()

	scraper.SetSteamApiKey(config.Constants.SteamDevAPIKey)
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
			return errors.New("models.UpdatePlayerInfo: " + hErr.Error())
		}

		player.GameHours = pHours
	}

	player.Profileurl = playerInfo.Profileurl
	player.Avatar = playerInfo.Avatar
	player.Name = playerInfo.Name

	return nil
}

func (player *Player) SetSetting(key string, value string) {
	if player.Settings == nil {
		player.Settings = make(gorm.Hstore)
	}

	player.Settings[key] = &value
	player.Save()
}

func (player *Player) GetSetting(key string) string {
	db.DB.First(player)
	if player.Settings == nil {
		return ""
	}

	value, ok := player.Settings[key]
	if !ok {
		return ""
	}

	return *value
}

func (t PlayerBanType) String() string {
	return map[PlayerBanType]string{
		PlayerBanJoin:   "lobby join ban",
		PlayerBanCreate: "lobby create ban",
		PlayerBanChat:   "chat ban",
		PlayerBanFull:   "full ban",
	}[t]
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
	// first check if player is already banned
	if banned := player.IsBanned(t); banned {
		db.DB.Model(&PlayerBan{}).Where("player_id = ? AND type = ? AND active = TRUE AND until > now()", player.ID, t).Update("until", tim)
		return nil
	}
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

var client = &http.Client{Timeout: 5 * time.Second}

//IsSubscribed returns whether if the player has subscribed to the given twitch channel.
//The player object should have a valid access token and twitch name
func (p *Player) IsSubscribed(channel string) bool {
	url := fmt.Sprintf("https://api.twitch.tv/kraken/users/%s/subscriptions/%s", p.TwitchName, channel)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/vnd.twitchtv.v3+json")
	req.Header.Add("Authorization", "OAuth "+p.TwitchAccessToken)

	resp, err := client.Do(req)
	if err != nil {
		logrus.Error(err)
		return false
	}

	var reply struct {
		ID string `json:"_id"`
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&reply)

	//if status code is 404, the user isn't subscribed
	return err == nil && resp.StatusCode != 404
}

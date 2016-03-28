package models

import (
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/jinzhu/gorm"
)

type PlayerBanType int

const (
	PlayerBanJoin PlayerBanType = iota
	PlayerBanCreate
	PlayerBanChat
	PlayerBanFull
	PlayerBanJoinMumble
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

func (t PlayerBanType) String() string {
	return map[PlayerBanType]string{
		PlayerBanJoin:       "lobby join ban",
		PlayerBanJoinMumble: "mumble lobby join ban",
		PlayerBanCreate:     "lobby create ban",
		PlayerBanChat:       "chat ban",
		PlayerBanFull:       "full ban",
	}[t]
}

func (player *Player) IsBannedWithTime(t PlayerBanType) (bool, time.Time) {
	ban := &PlayerBan{}
	err := db.DB.Where("type IN (?) AND until > now() AND player_id = ? AND active = TRUE", []PlayerBanType{t, PlayerBanFull}, player.ID).Order("until desc").First(ban).Error
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

func (player *Player) GetActiveBan(banType PlayerBanType) (*PlayerBan, error) {
	//try getting the full ban first
	ban := &PlayerBan{}
	err := db.DB.Model(&PlayerBan{}).Where("player_id = ? AND type = ? and active = TRUE", player.ID, PlayerBanFull).First(ban).Error

	if err != nil {
		err = db.DB.Model(&PlayerBan{}).Where("player_id = ? AND type = ? AND active = TRUE", player.ID, banType).First(ban).Error
		return ban, err
	}

	return ban, nil
}

func (player *Player) GetActiveBans() ([]*PlayerBan, error) {
	var bans []*PlayerBan
	err := db.DB.Where("player_id = ? AND active = TRUE AND until > now()", player.ID).Find(&bans).Error
	return bans, err
}

func (player *Player) GetAllBans() ([]*PlayerBan, error) {
	var bans []*PlayerBan
	err := db.DB.Where("player_id = ?", player.ID).Find(&bans).Error
	return bans, err

}

func GetAllActiveBans() []*PlayerBan {
	var bans []*PlayerBan
	db.DB.Where("active = TRUE AND until > now()").Find(&bans)
	return bans
}

func GetAllBans() []*PlayerBan {
	var bans []*PlayerBan
	db.DB.Model(&PlayerBan{}).Find(&bans)
	return bans
}

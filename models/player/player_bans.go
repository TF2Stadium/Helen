package player

import (
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/jinzhu/gorm"
)

type BanType int

const (
	BanJoin BanType = iota
	BanCreate
	BanChat
	BanFull
	BanJoinMumble
)

//PlayerBan represents a player ban
type Ban struct {
	gorm.Model
	PlayerID uint      // ID of the player banned
	Type     BanType   // Ban type
	Until    time.Time // Time until which the ban is valid
	Reason   string    // Reason for the ban
	Active   bool      `sql:"default:true"` // Whether the ban is active
}

func (t BanType) String() string {
	return map[BanType]string{
		BanJoin:       "lobby join ban",
		BanJoinMumble: "mumble lobby join ban",
		BanCreate:     "lobby create ban",
		BanChat:       "chat ban",
		BanFull:       "full ban",
	}[t]
}

func (player *Player) IsBannedWithTime(t BanType) (bool, time.Time) {
	ban := &Ban{}
	err := db.DB.Where("type IN (?) AND until > now() AND player_id = ? AND active = TRUE", []BanType{t, BanFull}, player.ID).Order("until desc").First(ban).Error
	if err != nil {
		return false, time.Time{}
	}

	return true, ban.Until
}

func (player *Player) IsBanned(t BanType) bool {
	res, _ := player.IsBannedWithTime(t)
	return res
}

func (player *Player) BanUntil(tim time.Time, t BanType, reason string) error {
	// first check if player is already banned
	if banned := player.IsBanned(t); banned {
		db.DB.Model(&Ban{}).Where("player_id = ? AND type = ? AND active = TRUE AND until > now()", player.ID, t).Update("until", tim)
		return nil
	}
	ban := Ban{
		PlayerID: player.ID,
		Type:     t,
		Until:    tim,
		Reason:   reason,
	}

	return db.DB.Create(&ban).Error
}

func (player *Player) Unban(t BanType) error {
	return db.DB.Model(&Ban{}).Where("player_id = ? AND type = ? AND active = TRUE", player.ID, t).
		Update("active", "FALSE").Error
}

func (player *Player) GetActiveBan(banType BanType) (*Ban, error) {
	//try getting the full ban first
	ban := &Ban{}
	err := db.DB.Model(&Ban{}).Where("player_id = ? AND type = ? and active = TRUE", player.ID, BanFull).First(ban).Error

	if err != nil {
		err = db.DB.Model(&Ban{}).Where("player_id = ? AND type = ? AND active = TRUE", player.ID, banType).First(ban).Error
		return ban, err
	}

	return ban, nil
}

func (player *Player) GetActiveBans() ([]*Ban, error) {
	var bans []*Ban
	err := db.DB.Where("player_id = ? AND active = TRUE AND until > now()", player.ID).Find(&bans).Error
	return bans, err
}

func (player *Player) GetAllBans() ([]*Ban, error) {
	var bans []*Ban
	err := db.DB.Where("player_id = ?", player.ID).Find(&bans).Error
	return bans, err

}

func GetAllActiveBans() []*Ban {
	var bans []*Ban
	db.DB.Where("active = TRUE AND until > now()").Find(&bans)
	return bans
}

func GetAllBans() []*Ban {
	var bans []*Ban
	db.DB.Model(&Ban{}).Find(&bans)
	return bans
}

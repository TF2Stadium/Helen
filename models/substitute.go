package models

import (
	"time"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
)

type Substitute struct {
	ID        uint      `gorm:"primary_key"json:"-"`
	CreatedAt time.Time `json:"-"`

	SteamID string `json:"-"`
	Filled  bool   `json:"-"`

	LobbyID uint   `json:"id"`
	Format  string `json:"type"`
	MapName string `json:"map"`
	Region  string `json:"region"`
	Mumble  bool   `json:"mumbleRequired"`

	Team  string `json:"team"`
	Class string `json:"class"`
}

func NewSub(lobbyid uint, steamid string) (*Substitute, error) {
	player, err := GetPlayerBySteamId(steamid)
	if err != nil {
		return nil, err
	}

	//helpers.Logger.Debug("#%d: Reported player %s<%s>",
	//	lobbyid, player.Name, player.SteamId)
	lob, _ := GetLobbyById(lobbyid)
	slot := &LobbySlot{}

	db.DB.Where("lobby_id = ? AND player_id = ?", lobbyid, player.ID).First(slot)

	sub := &Substitute{}

	sub.LobbyID = lob.ID
	sub.Format = FormatMap[lob.Type]
	sub.SteamID = player.SteamId
	sub.MapName = lob.MapName
	sub.Region = lob.RegionName
	sub.Mumble = lob.Mumble

	sub.Team = slot.Team
	sub.Class = slot.Class

	return sub, nil
}

func GetSubList() []*Substitute {
	var allSubs []*Substitute
	db.DB.Table("substitutes").Where("filled = ?", false).Find(&allSubs)

	return allSubs
}

func BroadcastSubList() {
	broadcaster.SendMessageToRoom("0_public", "subListData", GetSubList())
}

func GetPlayerSubs(steamid string) ([]*Substitute, error) {
	var subs []*Substitute

	err := db.DB.Table("substitutes").Where("steam_id = ?", steamid).Find(&subs).Error

	return subs, err
}

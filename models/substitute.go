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

	Slot int `json:"-"`
	// JSON stuff
	Team  string `slot:"-" json:"team"`
	Class string `slot:"-" json:"class"`
}

func NewSub(lobbyid uint, steamid string) (*Substitute, error) {
	player, err := GetPlayerBySteamID(steamid)
	if err != nil {
		return nil, err
	}

	lobby := &Lobby{}
	db.DB.First(lobby, lobbyid)
	//helpers.Logger.Debug("#%d: Reported player %s<%s>",
	//	lobbyid, player.Name, player.SteamId)
	lob, _ := GetLobbyByID(lobbyid)
	slot := &LobbySlot{}

	db.DB.Where("lobby_id = ? AND player_id = ?", lobbyid, player.ID).First(slot)

	sub := &Substitute{}

	sub.LobbyID = lob.ID
	sub.Format = formatMap[lob.Type]
	sub.SteamID = player.SteamID
	sub.MapName = lob.MapName
	sub.Region = lob.RegionName
	sub.Mumble = lob.Mumble

	sub.Slot = slot.Slot
	//sub.Team, sub.Class, _ = LobbyGetSlotInfoString(lobby.Type, slot.Slot)

	return sub, nil
}

func SubAndRemove(lobby *Lobby, player *Player) error {
	sub, err := NewSub(lobby.ID, player.SteamID)
	if err != nil {
		return err
	}

	db.DB.Save(sub)
	if tperr := lobby.RemovePlayer(player); tperr != nil {
		return tperr
	}
	BroadcastSubList()

	return nil
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

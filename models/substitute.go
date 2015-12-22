package models

import (
	"time"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
)

type Substitute struct {
	ID        uint      `gorm:"primary_key"json:"-"`
	CreatedAt time.Time `json:"-"`

	PlayerID uint `json:"-"`
	Filled   bool `json:"-"`

	LobbyID uint   `json:"id"`
	Format  string `json:"type"`
	MapName string `json:"map"`
	Region  string `json:"region"`
	Mumble  bool   `json:"mumbleRequired"`

	Slot int `json:"-"`
	// JSON stuff
	Team  string `sql:"-" json:"team"`
	Class string `sql:"-" json:"class"`
}

func NewSub(lobbyid, playerid uint) (*Substitute, error) {
	player := &Player{}
	err := db.DB.First(player, playerid).Error
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
	sub.PlayerID = player.ID
	sub.MapName = lob.MapName
	sub.Region = lob.RegionName
	sub.Mumble = lob.Mumble

	sub.Slot = slot.Slot
	//sub.Team, sub.Class, _ = LobbyGetSlotInfoString(lobby.Type, slot.Slot)

	return sub, nil
}

func (s *Substitute) Save() {
	db.DB.Save(s)
}

func SubAndRemove(lobby *Lobby, player *Player) error {
	sub, err := NewSub(lobby.ID, player.ID)
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

func GetAllSubs() []*Substitute {
	var allSubs []*Substitute
	db.DB.Table("substitutes").Where("filled = ?", false).Find(&allSubs)
	for _, sub := range allSubs {
		lobby, _ := GetLobbyByID(sub.LobbyID)

		sub.Team, sub.Class, _ = LobbyGetSlotInfoString(lobby.Type, sub.Slot)
	}

	return allSubs
}

func BroadcastSubList() {
	broadcaster.SendMessageToRoom("0_public", "subListData", GetAllSubs())
}

func GetPlayerSubs(player *Player) ([]*Substitute, error) {
	var subs []*Substitute

	err := db.DB.Table("substitutes").Where("player_id = ?", player.ID).Find(&subs).Error

	for _, sub := range subs {
		lobby, _ := GetLobbyByID(sub.LobbyID)

		sub.Team, sub.Class, _ = LobbyGetSlotInfoString(lobby.Type, sub.Slot)
	}

	return subs, err
}

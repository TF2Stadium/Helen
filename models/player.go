package models

import (
	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/jinzhu/gorm"
)

type Player struct {
	gorm.Model
	SteamId string `sql:"unique"` // Players steam ID
	Name    string // Player name
}

func NewPlayer(steamId string) *Player {
	player := &Player{SteamId: steamId}

	// magically get the player's name, avatar and other stuff from steam

	return player
}

func (player *Player) Save() error {
	var err error
	if db.DB.NewRecord(player) {
		err = db.DB.Create(player).Error
	} else {
		err = db.DB.Save(player).Error
	}
	return err
}

func GetPlayerBySteamId(steamid string) (*Player, *helpers.TPError) {
	var player = Player{}
	err := db.DB.Where("steam_id = ?", steamid).First(&player).Error
	if err != nil {
		return nil, helpers.NewTPError("Player is not in the database", -1)
	}
	return &player, nil
}

func (player *Player) GetLobbyId() (uint, error) {
	//TODO will need a few joins
	return 0, nil
}

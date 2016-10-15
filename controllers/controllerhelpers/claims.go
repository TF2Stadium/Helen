package controllerhelpers

import (
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models/player"
)

type TF2StadiumClaims struct {
	PlayerID       uint               `json:"player_id"`
	SteamID        string             `json:"steam_id"`
	MumblePassword string             `json:"mumble_password"`
	Role           authority.AuthRole `json:"role"`
	IssuedAt       int64              `json:"iat"`
	Issuer         string             `json:"iss"`
}

func playerExists(id uint, steamID string) bool {
	var count int
	db.DB.Model(&player.Player{}).Where("id = ? AND steam_id = ?", id, steamID).Count(count)
	return count != 0
}

func (c TF2StadiumClaims) Valid() error {
	if !playerExists(c.PlayerID, c.SteamID) {
		return player.ErrPlayerNotFound
	}

	return nil
}

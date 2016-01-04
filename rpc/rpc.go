package rpc

import (
	"errors"
	"net/rpc"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

type Helen struct{}

type Args struct {
	LobbyID uint
	Type    models.LobbyType

	SteamID string

	Team, Class string
}

func init() {
	rpc.Register(Helen{})
}

func getPlayerID(steamID string) uint {
	var id uint

	db.DB.DB().QueryRow("SELECT id FROM players WHERE steam_id = $1", steamID).Scan(&id)
	return id
}

// GetPlayerID returns a player ID (primary key), given their Steam Community id
func (Helen) GetPlayerID(steamID string, id *uint) error {
	*id = getPlayerID(steamID)
	return nil
}

func (Helen) GetTeam(args Args, team *string) error {
	var slot int

	db.DB.DB().QueryRow("SELECT slot FROM lobby_slots WHERE player_id = $1", getPlayerID(args.SteamID)).Scan(&slot)
	*team, _, _ = models.LobbyGetSlotInfoString(args.Type, slot)

	return nil
}

func (Helen) GetSteamIDFromSlot(args Args, steamID *string) error {
	slot, tperr := models.LobbyGetPlayerSlot(args.Type, args.Team, args.Class)
	var playerID uint

	if tperr != nil {
		return errors.New(tperr.Error())
	}

	err := db.DB.DB().QueryRow("SELECT player_id FROM lobby_slots WHERE slot = $1 AND lobby_id = $2", slot, args.LobbyID).Scan(&playerID)
	if err != nil {
		helpers.Logger.Error(err.Error())
	}

	err = db.DB.DB().QueryRow("SELECT steam_id FROM players WHERE id = $1", playerID).Scan(&steamID)
	if err != nil {
		helpers.Logger.Error(err.Error())
	}

	return nil
}

func (Helen) GetNameFromSteamID(steamID string, name *string) error {
	return db.DB.DB().QueryRow("SELECT name FROM players WHERE steam_id = $1", steamID).Scan(name)
}

func (Helen) IsAllowed(args Args, ok *bool) error {
	var count int
	playerID := getPlayerID(args.SteamID)

	db.DB.Table("lobby_slots").Where("lobby_id = ? AND player_id = ?", args.LobbyID, playerID).Count(&count)
	*ok = (count == 1)

	return nil
}

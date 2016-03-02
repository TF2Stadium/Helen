package handler

import (
	"errors"
	"regexp"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

type Player struct{}

func (Player) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Player) PlayerReady(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)
	lobbyid, tperr := player.GetLobbyID(false)
	if tperr != nil {
		return tperr
	}

	lobby, tperr := models.GetLobbyByIDServer(lobbyid)
	if tperr != nil {
		return tperr
	}

	if lobby.State != models.LobbyStateReadyingUp {
		return errors.New("Lobby hasn't been filled up yet.")
	}

	tperr = lobby.ReadyPlayer(player)

	if tperr != nil {
		return tperr
	}

	if lobby.IsEveryoneReady() {
		lobby.Start()

		hooks.BroadcastLobbyStart(lobby)
		models.BroadcastLobbyList()
	}

	return emptySuccess
}

func (Player) PlayerNotReady(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)
	lobbyid, tperr := player.GetLobbyID(false)
	if tperr != nil {
		return tperr
	}

	lobby, tperr := models.GetLobbyByID(lobbyid)
	if tperr != nil {
		return tperr
	}

	if lobby.State != models.LobbyStateReadyingUp {
		return errors.New("Lobby hasn't been filled up yet.")
	}

	tperr = lobby.UnreadyPlayer(player)
	lobby.RemovePlayer(player)
	hooks.AfterLobbyLeave(lobby, player)
	lobby.AddSpectator(player)
	hooks.AfterLobbySpec(socket.AuthServer, so, lobby)

	if tperr != nil {
		return tperr
	}

	lobby.UnreadyAllPlayers()
	return emptySuccess
}

func (Player) PlayerSettingsGet(so *wsevent.Client, args struct {
	Key *string `json:"key"`
}) interface{} {

	player := chelpers.GetPlayer(so.Token)
	if *args.Key == "*" {
		return newResponse(player.Settings)
	}

	setting := player.GetSetting(*args.Key)
	return newResponse(setting)
}

var reMumbleNick = regexp.MustCompile(`\w+`)

func (Player) PlayerSettingsSet(so *wsevent.Client, args struct {
	Key   *string `json:"key"`
	Value *string `json:"value"`
}) interface{} {

	player := chelpers.GetPlayer(so.Token)

	switch *args.Key {
	case "siteAlias":
		if len(*args.Value) > 32 {
			return errors.New("Site alias must be under 32 characters long.")
		}

		player.SetSetting(*args.Key, *args.Value)

		player.SetPlayerProfile()
		so.EmitJSON(helpers.NewRequest("playerProfile", player))

		if lobbyID, _ := player.GetLobbyID(true); lobbyID != 0 {
			lobby, _ := models.GetLobbyByID(lobbyID)
			lobbyData := lobby.LobbyData(true)
			lobbyData.Send()
		}
	case "mumbleNick":
		if !reMumbleNick.MatchString(*args.Value) {
			return errors.New("Invalid Mumble nick.")
		}
		if len(*args.Value) > 32 {
			return errors.New("Username is too long.")
		}

		if player.MumbleUsername == *args.Value {
			return emptySuccess
		} else if player.MumbleUsername == "" {
			return errors.New("Username cannot be empty.")
		}

		var count int
		db.DB.Model(&models.Player{}).Where("mumble_username = ?", *args.Value).Count(&count)
		if count != 0 {
			return errors.New("This username has already been taken.")
		}

		player.MumbleUsername = *args.Value
		player.Save()
	default:
		player.SetSetting(*args.Key, *args.Value)
	}

	return emptySuccess
}

func (Player) PlayerProfile(so *wsevent.Client, args struct {
	Steamid *string `json:"steamid"`
}) interface{} {

	steamid := *args.Steamid
	if steamid == "" {
		steamid = so.Token.Claims["steam_id"].(string)
	}

	player, err := models.GetPlayerBySteamID(steamid)
	if err != nil {
		return err
	}

	player.SetPlayerProfile()
	return newResponse(player)
}

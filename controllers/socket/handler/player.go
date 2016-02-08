package handler

import (
	"regexp"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
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
	steamid := chelpers.GetSteamId(so.ID)
	player, tperr := models.GetPlayerBySteamID(steamid)
	if tperr != nil {
		return tperr
	}

	lobbyid, tperr := player.GetLobbyID(false)
	if tperr != nil {
		return tperr
	}

	lobby, tperr := models.GetLobbyByIDServer(lobbyid)
	if tperr != nil {
		return tperr
	}

	if lobby.State != models.LobbyStateReadyingUp {
		return helpers.NewTPError("Lobby hasn't been filled up yet.", 4)
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

	return chelpers.EmptySuccessJS
}

func (Player) PlayerNotReady(so *wsevent.Client, _ struct{}) interface{} {
	player, tperr := models.GetPlayerBySteamID(chelpers.GetSteamId(so.ID))

	if tperr != nil {
		return tperr
	}

	lobbyid, tperr := player.GetLobbyID(false)
	if tperr != nil {
		return tperr
	}

	lobby, tperr := models.GetLobbyByID(lobbyid)
	if tperr != nil {
		return tperr
	}

	if lobby.State != models.LobbyStateReadyingUp {
		return helpers.NewTPError("Lobby hasn't been filled up yet.", 4)
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
	return chelpers.EmptySuccessJS
}

func (Player) PlayerSettingsGet(so *wsevent.Client, args struct {
	Key *string `json:"key"`
}) interface{} {

	player, _ := models.GetPlayerBySteamID(chelpers.GetSteamId(so.ID))

	if *args.Key == "*" {
		return chelpers.NewResponse(player.Settings)
	}

	setting := player.GetSetting(*args.Key)
	return chelpers.NewResponse(setting)
}

var reMumbleNick = regexp.MustCompile(`\w+`)

func (Player) PlayerSettingsSet(so *wsevent.Client, args struct {
	Key   *string `json:"key"`
	Value *string `json:"value"`
}) interface{} {

	player, _ := models.GetPlayerBySteamID(chelpers.GetSteamId(so.ID))

	player.SetSetting(*args.Key, *args.Value)

	switch *args.Key {
	case "siteAlias":
		profile := models.DecoratePlayerProfileJson(player)
		so.EmitJSON(helpers.NewRequest("playerProfile", profile))

		if lobbyID, _ := player.GetLobbyID(true); lobbyID != 0 {
			lobby, _ := models.GetLobbyByID(lobbyID)
			lobbyData := lobby.LobbyData(true)
			lobbyData.Send()
		}
	case "mumbleNick":
		if !reMumbleNick.MatchString(*args.Value) {
			return helpers.NewTPError("Invalid Mumble nick.", -1)
		}
	}

	return chelpers.EmptySuccessJS
}

func (Player) PlayerProfile(so *wsevent.Client, args struct {
	Steamid *string `json:"steamid"`
}) interface{} {

	steamid := *args.Steamid
	if steamid == "" {
		steamid = chelpers.GetSteamId(so.ID)
	}

	player, playErr := models.GetPlayerWithStats(steamid)

	if playErr != nil {
		return playErr
	}

	result := models.DecoratePlayerProfileJson(player)
	return chelpers.NewResponse(result)
}

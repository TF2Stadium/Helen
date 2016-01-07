package handler

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

type Player struct{}

func (Player) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Player) PlayerReady(_ *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	steamid := chelpers.GetSteamId(so.Id())
	player, tperr := models.GetPlayerBySteamID(steamid)
	if tperr != nil {
		return tperr
	}

	lobbyid, tperr := player.GetLobbyID(false)
	if tperr != nil {
		return tperr
	}

	lobby, tperr := models.GetLobbyByIdServer(lobbyid)
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

		chelpers.BroadcastLobbyStart(lobby)
		models.BroadcastLobbyList()
		models.FumbleLobbyStarted(lobby)
	}

	return chelpers.EmptySuccessJS
}

func (Player) PlayerNotReady(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	player, tperr := models.GetPlayerBySteamID(chelpers.GetSteamId(so.Id()))

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
	chelpers.AfterLobbyLeave(server, so, lobby, player)

	if tperr != nil {
		return tperr
	}

	lobby.UnreadyAllPlayers()
	return chelpers.EmptySuccessJS
}

func (Player) PlayerSettingsGet(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	var args struct {
		Key *string `json:"key"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	player, _ := models.GetPlayerBySteamID(chelpers.GetSteamId(so.Id()))

	var settings []models.PlayerSetting
	var setting models.PlayerSetting
	if *args.Key == "*" {
		settings, err = player.GetSettings()
	} else {
		setting, err = player.GetSetting(*args.Key)
		settings = append(settings, setting)
	}

	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	result := models.DecoratePlayerSettingsJson(settings)
	return chelpers.BuildSuccessJSON(result)
}

func (Player) PlayerSettingsSet(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	var args struct {
		Key   *string `json:"key"`
		Value *string `json:"value"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	player, _ := models.GetPlayerBySteamID(chelpers.GetSteamId(so.Id()))

	err = player.SetSetting(*args.Key, *args.Value)
	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	if *args.Key == "siteAlias" {
		profile := models.DecoratePlayerProfileJson(player)
		so.EmitJSON(helpers.NewRequest("playerProfile", profile))

		if lobbyID, _ := player.GetLobbyID(true); lobbyID != 0 {
			lobby, _ := models.GetLobbyByID(lobbyID)
			lobbyData := lobby.LobbyData(true)
			lobbyData.Send()
		}
	}

	return chelpers.EmptySuccessJS
}

func (Player) PlayerProfile(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	var args struct {
		Steamid *string `json:"steamid"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	steamid := *args.Steamid
	if steamid == "" {
		steamid = chelpers.GetSteamId(so.Id())
	}

	player, playErr := models.GetPlayerWithStats(steamid)

	if playErr != nil {
		return playErr
	}

	result := models.DecoratePlayerProfileJson(player)
	return chelpers.BuildSuccessJSON(result)
}

package handler

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

type Player struct{}

func (Player) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Player) PlayerReady(_ *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr.Encode()
	}

	steamid := chelpers.GetSteamId(so.Id())
	player, tperr := models.GetPlayerBySteamId(steamid)
	if tperr != nil {
		return tperr.Encode()
	}

	lobbyid, tperr := player.GetLobbyId()
	if tperr != nil {
		return tperr.Encode()
	}

	lobby, tperr := models.GetLobbyByIdServer(lobbyid)
	if tperr != nil {
		return tperr.Encode()
	}

	if lobby.State != models.LobbyStateReadyingUp {
		return helpers.NewTPError("Lobby hasn't been filled up yet.", 4).Encode()
	}

	tperr = lobby.ReadyPlayer(player)

	if tperr != nil {
		return tperr.Encode()
	}

	if lobby.IsEveryoneReady() {
		mapLock.Lock()
		timeoutStop[lobby.ID] <- struct{}{}
		close(timeoutStop[lobby.ID])
		delete(timeoutStop, lobby.ID)
		mapLock.Unlock()
		lobby.State = models.LobbyStateInProgress
		lobby.Save()

		chelpers.BroadcastLobbyStart(lobby)
		models.BroadcastLobbyList()
		models.FumbleLobbyStarted(lobby)
	}

	return chelpers.EmptySuccessJS
}

func (Player) PlayerNotReady(_ *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, authority.AuthAction(0), true)

	if reqerr != nil {
		return reqerr.Encode()
	}

	player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	if tperr != nil {
		return tperr.Encode()
	}

	lobbyid, tperr := player.GetLobbyId()
	if tperr != nil {
		return tperr.Encode()
	}

	lobby, tperr := models.GetLobbyById(lobbyid)
	if tperr != nil {
		return tperr.Encode()
	}

	if lobby.State != models.LobbyStateReadyingUp {
		return helpers.NewTPError("Lobby hasn't been filled up yet.", 4).Encode()
	}

	tperr = lobby.UnreadyPlayer(player)
	lobby.RemovePlayer(player)

	if tperr != nil {
		return tperr.Encode()
	}

	lobby.UnreadyAllPlayers()
	mapLock.Lock()
	c, ok := timeoutStop[lobby.ID]
	if ok {
		close(c)
		delete(timeoutStop, lobby.ID)
	}
	mapLock.Unlock()

	return chelpers.EmptySuccessJS
}

func (Player) PlayerSettingsGet(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}
	var args struct {
		Key string `json:"key"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	var settings []models.PlayerSetting
	var setting models.PlayerSetting
	if args.Key == "*" {
		settings, err = player.GetSettings()
	} else {
		setting, err = player.GetSetting(args.Key)
		settings = append(settings, setting)
	}

	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	result := models.DecoratePlayerSettingsJson(settings)
	resp, _ := chelpers.BuildSuccessJSON(result).Encode()
	return resp
}

func (Player) PlayerSettingsSet(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}
	var args struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	err = player.SetSetting(args.Key, args.Value)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	return chelpers.EmptySuccessJS
}

func (Player) PlayerProfile(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		return reqerr.Encode()
	}
	var args struct {
		Steamid string `json:"steamid"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	steamid := args.Steamid
	if steamid == "" {
		steamid = chelpers.GetSteamId(so.Id())
	}

	player, playErr := models.GetPlayerWithStats(steamid)

	if playErr != nil {
		return playErr.Encode()
	}

	result := models.DecoratePlayerProfileJson(player)
	resp, _ := chelpers.BuildSuccessJSON(result).Encode()
	return resp
}

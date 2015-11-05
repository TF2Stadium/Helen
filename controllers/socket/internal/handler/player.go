package handler

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/vibhavp/wsevent"
)

func PlayerSettingsGet(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}
	var args struct {
		Key string `json:"key"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
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
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), 0).Encode()
		return string(bytes)
	}

	result := models.DecoratePlayerSettingsJson(settings)
	resp, _ := chelpers.BuildSuccessJSON(result).Encode()
	return string(resp)
}

func PlayerSettingsSet(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}
	var args struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	err = player.SetSetting(args.Key, args.Value)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), 0).Encode()
		return string(bytes)
	}

	resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
	return string(resp)
}

func PlayerProfile(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := reqerr.ErrorJSON().Encode()
		return string(bytes)
	}
	var args struct {
		Steamid string `json:"steamid"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := chelpers.BuildFailureJSON(err.Error(), -1).Encode()
		return string(bytes)
	}

	steamid := args.Steamid
	if steamid == "" {
		steamid = chelpers.GetSteamId(so.Id())
	}

	player, playErr := models.GetPlayerWithStats(steamid)

	if playErr != nil {
		bytes, _ := chelpers.BuildFailureJSON(playErr.Error(), 0).Encode()
		return string(bytes)
	}

	result := models.DecoratePlayerProfileJson(player)
	resp, _ := chelpers.BuildSuccessJSON(result).Encode()
	return string(resp)
}

package handler

import (
	"encoding/json"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
	"github.com/bitly/go-simplejson"
)

func PlayerSettingsGet(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}
	var args struct {
		Key string `json:"key"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
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
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	result := models.DecoratePlayerSettingsJson(settings)
	resp, _ := chelpers.BuildSuccessJSON(result).Encode()
	return string(resp)
}

func PlayerSettingsSet(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}
	var args struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

	err = player.SetSetting(args.Key, args.Value)
	if err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	return chelpers.EmptySuccessJS
}

func PlayerProfile(server *wsevent.Server, so *wsevent.Client, data string) string {
	reqerr := chelpers.FilterRequest(so, 0, true)

	if reqerr != nil {
		bytes, _ := json.Marshal(reqerr)
		return string(bytes)
	}
	var args struct {
		Steamid string `json:"steamid"`
	}

	err := chelpers.GetParams(data, &args)
	if err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	steamid := args.Steamid
	if steamid == "" {
		steamid = chelpers.GetSteamId(so.Id())
	}

	player, playErr := models.GetPlayerWithStats(steamid)

	if playErr != nil {
		bytes, _ := playErr.Encode()
		return string(bytes)
	}

	result := models.DecoratePlayerProfileJson(player)
	resp, _ := chelpers.BuildSuccessJSON(result).Encode()
	return string(resp)
}

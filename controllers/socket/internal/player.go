package handler

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
	"reflect"
)

var playerSettingsGetFilter = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"key": chelpers.Param{Kind: reflect.String, Default: ""},
	},
}

func PlayerSettingsGet(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, playerSettingsGetFilter,
		func(params map[string]interface{}) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			key := params["key"].(string)

			var err error
			var settings []models.PlayerSetting
			var setting models.PlayerSetting
			if key == "" {
				settings, err = player.GetSettings()
			} else {
				setting, err = player.GetSetting(key)
				settings = append(settings, setting)
			}

			if err != nil {
				bytes, _ := chelpers.BuildFailureJSON(err.Error(), 0).Encode()
				return string(bytes)
			}

			result := models.DecoratePlayerSettingsJson(settings)
			resp, _ := chelpers.BuildSuccessJSON(result).Encode()
			return string(resp)
		})
}

var playerSettingsSetFilter = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"key":   chelpers.Param{Kind: reflect.String},
		"value": chelpers.Param{Kind: reflect.String},
	},
}

func PlayerSettingsSet(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, playerSettingsSetFilter,
		func(params map[string]interface{}) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			key := params["key"].(string)
			value := params["value"].(string)

			err := player.SetSetting(key, value)
			if err != nil {
				bytes, _ := chelpers.BuildFailureJSON(err.Error(), 0).Encode()
				return string(bytes)
			}

			resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(resp)
		})
}

var playerProfileFilter = chelpers.FilterParams{
	FilterLogin: true,
	Params: map[string]chelpers.Param{
		"steamid": chelpers.Param{Kind: reflect.String, Default: ""},
	},
}

func PlayerProfile(so socketio.Socket) func(string) string {
	return chelpers.FilterRequest(so, playerProfileFilter,
		func(params map[string]interface{}) string {

			steamid := params["steamid"].(string)

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
		})
}

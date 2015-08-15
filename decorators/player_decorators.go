package decorators

import (
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
)

func GetPlayerSettingsJson(settings []models.PlayerSetting) *simplejson.Json {
	json := simplejson.New()

	for _, obj := range settings {
		json.Set(obj.Key, obj.Value)
	}

	return json
}

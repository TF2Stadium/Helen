package models_test

import (
	"fmt"
	"testing"

	"github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

var testSettingsData = []byte(`
{
	"formats": [
		{
			"name": "sixes",
			"prettyName": "6v6",
			"important": true
		},{
			"name": "highlander",
			"prettyName": "Highlander",
			"important": true
		},{
			"name": "fours",
			"prettyName": "4v4"
		}
	],
	"maps": [
		{
			"name": "cp_process_final",
			"formats": {
				"highlander": 1,
				"sixes": 2
			}
		},
		{
			"name": "pl_upward",
			"formats": {
				"highlander": 2
			}
		}
	],
	"leagues": [
		{
			"name": "etf2l",
			"prettyName": "ETF2L",
			"descriptions": {
				"cp": "Somethings cool happen"
			},
			"formats": {
				"highlander": true,
				"sixes": true
			}
		}
	],
	"whitelists": [
		{
			"id": 3250,
			"prettyName": "ETF2L Highlander (Season 8)",
			"league": "etf2l",
			"format": "highlander"
		}
	]
}`)

func TestSettingsLoad(t *testing.T) {
	assert := assert.New(t)

	err := models.LoadLobbySettings(testSettingsData)

	if assert.Nil(err) {
		if assert.Equal(len(models.LobbyFormats), 3) {
			assert.Equal(models.LobbyFormats[0].Name, "sixes")
			assert.Equal(models.LobbyFormats[0].PrettyName, "6v6")
			assert.Equal(models.LobbyFormats[0].Important, true)

			assert.Equal(models.LobbyFormats[1].Name, "highlander")
			assert.Equal(models.LobbyFormats[1].PrettyName, "Highlander")
			assert.Equal(models.LobbyFormats[1].Important, true)

			assert.Equal(models.LobbyFormats[2].Name, "fours")
			assert.Equal(models.LobbyFormats[2].PrettyName, "4v4")
			assert.Equal(models.LobbyFormats[2].Important, false)
		}

		if assert.Equal(len(models.LobbyMaps), 2) {
			assert.Equal(models.LobbyMaps[0].Name, "cp_process_final")
			if assert.Equal(len(models.LobbyMaps[0].Formats), 2) {
				assert.Equal(models.LobbyMaps[0].Formats[0].Format.Name, "highlander")
				assert.Equal(models.LobbyMaps[0].Formats[0].Importance, 1)

				assert.Equal(models.LobbyMaps[0].Formats[1].Format.Name, "sixes")
				assert.Equal(models.LobbyMaps[0].Formats[1].Importance, 2)
			}

			assert.Equal(models.LobbyMaps[1].Name, "pl_upward")
			if assert.Equal(len(models.LobbyMaps[1].Formats), 1) {
				assert.Equal(models.LobbyMaps[1].Formats[0].Format.Name, "highlander")
				assert.Equal(models.LobbyMaps[1].Formats[0].Importance, 2)
			}
		}

		// TODO write more tests, pls

		output, err := models.LobbySettingsToJson().Encode()
		if assert.Nil(err) {
			fmt.Println(string(output))
		}
	}
}

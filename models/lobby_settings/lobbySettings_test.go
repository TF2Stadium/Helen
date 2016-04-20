package lobbySettings_test

import (
	"testing"

	. "github.com/TF2Stadium/Helen/models/lobby_settings"
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

	err := LoadLobbySettings(testSettingsData)

	if assert.Nil(err) {
		// test formats
		if assert.Equal(3, len(LobbyFormats)) {
			if format, ok := GetLobbyFormat("sixes"); assert.True(ok) {
				assert.Equal("6v6", format.PrettyName)
				assert.Equal(true, format.Important)
			}

			if format, ok := GetLobbyFormat("highlander"); assert.True(ok) {
				assert.Equal("Highlander", format.PrettyName)
				assert.Equal(true, format.Important)
			}

			if format, ok := GetLobbyFormat("fours"); assert.True(ok) {
				assert.Equal("4v4", format.PrettyName)
				assert.Equal(false, format.Important)
			}
		}

		// test maps
		if assert.Equal(2, len(LobbyMaps)) {
			if amap, ok := GetLobbyMap("cp_process_final"); assert.True(ok) {
				assert.Equal(2, len(amap.Formats))

				if mapFormat, ok := amap.GetFormat("highlander"); assert.True(ok) {
					assert.Equal(1, mapFormat.Importance)
				}
				if mapFormat, ok := amap.GetFormat("sixes"); assert.True(ok) {
					assert.Equal(2, mapFormat.Importance)
				}
				if mapFormat, ok := amap.GetFormat("fours"); assert.True(ok) {
					assert.Equal(0, mapFormat.Importance)
				}
			}

			if amap, ok := GetLobbyMap("pl_upward"); assert.True(ok) {
				assert.Equal(1, len(amap.Formats))

				if mapFormat, ok := amap.GetFormat("highlander"); assert.True(ok) {
					assert.Equal(2, mapFormat.Importance)
				}
			}
		}

		// TODO write more tests, pls

		_, err := LobbySettingsToJSON().Encode()
		assert.NoError(err)
	}
}

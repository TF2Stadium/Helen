package assets

import _ "embed"

//go:embed lobbySettingsData.json
var LobbySettingsJSON []byte

//go:embed geoip.mmdb
var GeoIPDB []byte

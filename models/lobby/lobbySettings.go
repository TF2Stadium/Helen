// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package lobby

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TF2Stadium/Helen/assets"
	"github.com/bitly/go-simplejson"
)

type LobbyFormat struct {
	//gorm.Model -- not in a database yet
	Name       string
	PrettyName string
	Important  bool
}

type LobbyMapFormat struct {
	//gorm.Model -- not in a database yet
	Format     *LobbyFormat
	Importance int
}

type LobbyMap struct {
	//gorm.Model -- not in a database yet
	Name    string
	Formats []*LobbyMapFormat `gorm:"many2many:lobby_map_formats"`
}

func (m *LobbyMap) GetFormat(formatName string) (*LobbyMapFormat, bool) {
	for _, mapFormat := range m.Formats {
		if mapFormat.Format.Name == formatName {
			return mapFormat, true
		}
	}
	if format, ok := GetLobbyFormat(formatName); ok {
		return &LobbyMapFormat{
			Format:     format,
			Importance: 0,
		}, true
	}
	return nil, false
}

// TODO make int?
type MapType string

type LobbyLeagueDescription struct {
	//gorm.Model -- not in a database yet
	MapType     MapType
	Description string
}

type LobbyLeagueFormat struct {
	//gorm.Model -- not in a database yet
	Format *LobbyFormat
	Used   bool
}

type LobbyLeague struct {
	//gorm.Model -- not in a database yet
	Name         string
	PrettyName   string
	Descriptions []*LobbyLeagueDescription `gorm:"many2many:lobby_league_descriptions"`
	Formats      []*LobbyLeagueFormat      `gorm:"many2many:lobby_league_formats"`
}

type LobbyWhitelist struct {
	//gorm.Model -- not in a database yet
	ID         int
	PrettyName string
	League     *LobbyLeague
	Format     *LobbyFormat
}

var LobbyFormats []LobbyFormat
var lobbyFormatFromName map[string]int

var LobbyMaps []LobbyMap
var lobbyMapFromName map[string]int

var LobbyLeagues []LobbyLeague
var lobbyLeagueFromName map[string]int

var LobbyWhitelists []LobbyWhitelist
var lobbyWhitelistFromID map[int]int

func GetLobbyFormat(formatName string) (*LobbyFormat, bool) {
	if format, ok := lobbyFormatFromName[formatName]; ok {
		return &LobbyFormats[format], true
	}
	return nil, false
}

func GetLobbyMap(mapName string) (*LobbyMap, bool) {
	if amap, ok := lobbyMapFromName[mapName]; ok {
		return &LobbyMaps[amap], true
	}
	return nil, false
}

func GetLobbyLeague(leagueName string) (*LobbyLeague, bool) {
	if league, ok := lobbyLeagueFromName[leagueName]; ok {
		return &LobbyLeagues[league], true
	}
	return nil, false
}

func GetLobbyWhitelist(whitelistId int) (*LobbyWhitelist, bool) {
	if whitelist, ok := lobbyWhitelistFromID[whitelistId]; ok {
		return &LobbyWhitelists[whitelist], true
	}
	return nil, false
}

func LoadLobbySettingsFromFile(fileName string) error {
	data := assets.MustAsset(fileName)
	return LoadLobbySettings(data)
}

func LoadLobbySettings(data []byte) error {
	var args struct {
		Formats []struct {
			Name       string `json:"name"`
			PrettyName string `json:"prettyName"`
			Important  bool   `json:"important"`
		} `json:"formats"`
		Maps []struct {
			Name    string         `json:"name"`
			Formats map[string]int `json:"formats"`
		} `json:"maps"`
		Leagues []struct {
			Name         string            `json:"name"`
			PrettyName   string            `json:"prettyName"`
			Descriptions map[string]string `json:"descriptions"`
			Formats      map[string]bool   `json:"formats"`
		} `json:"leagues"`
		Whitelists []struct {
			ID         int    `json:"id"`
			PrettyName string `json:"prettyName"`
			League     string `json:"league"`
			Format     string `json:"format"`
		} `json:"whitelists"`
	}

	err := json.Unmarshal(data, &args)
	if err != nil {
		return err
	}

	// formats
	LobbyFormats = make([]LobbyFormat, len(args.Formats))
	lobbyFormatFromName = make(map[string]int)
	for i, format := range args.Formats {
		LobbyFormats[i] = LobbyFormat{
			Name:       format.Name,
			PrettyName: format.PrettyName,
			Important:  format.Important,
		}
		lobbyFormatFromName[format.Name] = i
	}

	// maps
	LobbyMaps = make([]LobbyMap, len(args.Maps))
	lobbyMapFromName = make(map[string]int)
	for i, amap := range args.Maps {
		lobbyMap := LobbyMap{
			Name:    amap.Name,
			Formats: make([]*LobbyMapFormat, 0, len(amap.Formats)),
		}
		for name, importance := range amap.Formats {
			if lobbyFormat, ok := GetLobbyFormat(name); ok {
				lobbyMap.Formats = append(lobbyMap.Formats, &LobbyMapFormat{
					Format:     lobbyFormat,
					Importance: importance,
				})
			} else {
				return errors.New(fmt.Sprintf("Referenced a non existing format %q", name))
			}
		}

		LobbyMaps[i] = lobbyMap
		lobbyMapFromName[amap.Name] = i
	}

	// leagues
	LobbyLeagues = make([]LobbyLeague, len(args.Leagues))
	lobbyLeagueFromName = make(map[string]int)
	for i, league := range args.Leagues {
		lobbyLeague := LobbyLeague{
			Name:         league.Name,
			PrettyName:   league.PrettyName,
			Descriptions: make([]*LobbyLeagueDescription, 0, len(league.Descriptions)),
			Formats:      make([]*LobbyLeagueFormat, 0, len(league.Formats)),
		}
		for atype, description := range league.Descriptions {
			lobbyLeagueDescription := &LobbyLeagueDescription{
				MapType:     MapType(atype),
				Description: description,
			}
			lobbyLeague.Descriptions = append(lobbyLeague.Descriptions, lobbyLeagueDescription)
		}
		for name, used := range league.Formats {
			if lobbyFormat, ok := GetLobbyFormat(name); ok {
				lobbyLeagueFormat := &LobbyLeagueFormat{
					Format: lobbyFormat,
					Used:   used,
				}
				lobbyLeague.Formats = append(lobbyLeague.Formats, lobbyLeagueFormat)
			} else {
				return errors.New(fmt.Sprintf("Referenced a non existing format %q", name))
			}
		}

		LobbyLeagues[i] = lobbyLeague
		lobbyLeagueFromName[league.Name] = i
	}

	// whitelists
	LobbyWhitelists = make([]LobbyWhitelist, len(args.Whitelists))
	lobbyWhitelistFromID = make(map[int]int)
	for i, whitelist := range args.Whitelists {
		if lobbyLeague, ok := GetLobbyLeague(whitelist.League); ok {
			if lobbyFormat, ok := GetLobbyFormat(whitelist.Format); ok {
				lobbyWhitelist := LobbyWhitelist{
					ID:         whitelist.ID,
					PrettyName: whitelist.PrettyName,
					League:     lobbyLeague,
					Format:     lobbyFormat,
				}

				LobbyWhitelists[i] = lobbyWhitelist
				lobbyWhitelistFromID[whitelist.ID] = i
			} else {
				return errors.New(fmt.Sprintf("Referenced a non existing format %q", whitelist.Format))
			}
		} else {
			return errors.New(fmt.Sprintf("Referenced a non existing league %q", whitelist.League))
		}
	}

	return nil
}

func LobbySettingsToJSON() *simplejson.Json {
	j := simplejson.New()

	// formats
	{
		formats := simplejson.New()

		formatList := make([]*simplejson.Json, len(LobbyFormats))
		for i, format := range LobbyFormats {
			f := simplejson.New()
			f.Set("value", format.Name)
			f.Set("title", format.PrettyName)
			f.Set("important", format.Important)

			formatList[i] = f
		}
		formats.Set("key", "type")
		formats.Set("title", "Format")
		formats.Set("options", formatList)

		j.Set("formats", formats)
	}

	// maps
	{
		maps := simplejson.New()

		mapList := make([]*simplejson.Json, len(LobbyMaps))
		for i, amap := range LobbyMaps {
			f := simplejson.New()
			f.Set("value", amap.Name)
			for _, mapFormat := range amap.Formats {
				f.Set(mapFormat.Format.Name, mapFormat.Importance)
			}

			mapList[i] = f
		}
		maps.Set("key", "mapName")
		maps.Set("title", "Map")
		maps.Set("options", mapList)

		j.Set("maps", maps)
	}

	// leagues
	{
		leagues := simplejson.New()

		leagueList := make([]*simplejson.Json, len(LobbyLeagues))
		for i, league := range LobbyLeagues {
			leagueDescs := simplejson.New()
			for _, leagueDesc := range league.Descriptions {
				leagueDescs.Set(string(leagueDesc.MapType), leagueDesc.Description)
			}

			f := simplejson.New()
			f.Set("value", league.Name)
			f.Set("title", league.PrettyName)
			f.Set("descriptions", leagueDescs)
			for _, leagueFormat := range league.Formats {
				f.Set(leagueFormat.Format.Name, leagueFormat.Used)
			}

			leagueList[i] = f
		}
		leagues.Set("key", "league")
		leagues.Set("title", "League")
		leagues.Set("options", leagueList)

		j.Set("leagues", leagues)
	}

	// whitelists
	{
		whitelists := simplejson.New()

		whitelistList := make([]*simplejson.Json, len(LobbyWhitelists))
		for i, whitelist := range LobbyWhitelists {
			f := simplejson.New()
			f.Set("value", whitelist.ID)
			f.Set("title", whitelist.PrettyName)
			f.Set("league", whitelist.League.Name)
			f.Set("format", whitelist.Format.Name)

			whitelistList[i] = f
		}
		whitelists.Set("key", "whitelist")
		whitelists.Set("title", "Whitelist")
		whitelists.Set("options", whitelistList)

		j.Set("whitelists", whitelists)
	}

	return j
}

package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/TeamPlayTF/Server/config"
)

const (
	ConfigsPath = "/configs/"
	ConfigsFile = "configs.json"
	MapsFile    = "maps.json"
)

var ConfigsData *ConfigData

// MapsData holds the map + config list from maps.json
//
// think about it like this:
// -> MapsData["MAP NAME"].V6.Ugc -> would return the ugc's 6v6 cfg for that map
// -> MapsData[string].GameMode.LeagueConfig
var MapsData map[string]GameMode

type LeagueConfig struct {
	Ugc   string `json:"ugc"`
	Etf2l string `json:"etf2l"`
}

type GameMode struct {
	V6 *LeagueConfig `json:"6v6"`
	V9 *LeagueConfig `json:"9v9"`
}

// league configs from configs.json
type ConfigData struct {
	Etf2l struct {
		V6 []string `json:"6v6"`
		V9 []string `json:"9v9"`
	} `json:"etf2l"`

	Ugc struct {
		V6 []string `json:"6v6"`
		V9 []string `json:"9v9"`
	} `json:"ugc"`
}

// configs.json
type ServerConfig struct {
	Name   string    // example: HL_stopwatch
	League string    // ugc, etf2l...
	Type   LobbyType // 9v9, 6v6...
	Data   string    // config file's text
	Map    string
}

// maps.json
type MapConfig struct {
	Name   string        // map name
	Config *ServerConfig // config info (name, league, type and data)
}

func InitConfigs() {
	// configs
	fmt.Println("[Configs.Init] Loading server configs...")
	cfgFile, cfgErr := ioutil.ReadFile(config.Constants.StaticFileLocation + ConfigsPath + ConfigsFile)

	if cfgErr == nil {
		json.Unmarshal(cfgFile, &ConfigsData)
		fmt.Println("[Configs.Init] Server configs loaded!")

	} else {
		fmt.Println("[Configs.Init] ERROR while trying to load server configs!")
		log.Fatal(cfgErr)
	}

	// maps
	fmt.Println("[Configs.Init] Loading maps configs...")
	mapFile, mapErr := ioutil.ReadFile(config.Constants.StaticFileLocation + ConfigsPath + MapsFile)

	if mapErr == nil {
		json.Unmarshal(mapFile, &MapsData)
		fmt.Println("[Configs.Init] Maps configs loaded!")

	} else {
		fmt.Println("[Configs.Init] ERROR while trying to load maps configs!")
		log.Fatal(mapErr)
	}
}

func NewServerConfig() *ServerConfig {
	return new(ServerConfig)
}

func (c *ServerConfig) Get() (string, error) {
	if c.League == "" {
		return "", errors.New("[Configs.Get]: No league specified!")
	}

	if c.Type != LobbyTypeSixes && c.Type != LobbyTypeHighlander {
		return "", errors.New("[Configs.Get]: The type you specified doesn't exists!")
	}

	if c.Name == "" {
		configName, configNameErr := c.GetMapConfig(c.Map)

		// if happens, rip
		if configNameErr != nil {
			log.Fatal(configNameErr)
		}

		fmt.Println("[Configs.Get]: Map config choosen: " + configName)

		if configName == "" {
			return "", errors.New("[Configs.Get]: No config name or map specified!")
		} else {
			c.Name = configName
		}
	}

	// get config's name
	cfgName, nameErr := c.GetName()

	fmt.Println("[Configs.Get]: Config that will be used: " + cfgName)

	if nameErr != nil {
		return "", nameErr
	}

	// gets the base config for each league
	// "the config that needs to run before the map type config"
	var preConfigName string
	var etf2lPreConfig []byte
	var etf2lPreErr error

	// etf2l
	if c.League == "etf2l" {
		preConfigName = "/etf2l/etf2l.cfg"

		var etf2lPreConfigName string
		if c.Type == LobbyTypeSixes {
			etf2lPreConfigName = "/etf2l/etf2l_6v6.cfg"

		} else if c.Type == LobbyTypeHighlander {
			etf2lPreConfigName = "/etf2l/etf2l_9v9.cfg"
		}

		// etf2l pre configs's pre config lol
		etf2lPreConfig, etf2lPreErr = ioutil.ReadFile(filepath.Clean(config.Constants.StaticFileLocation +
			ConfigsPath + etf2lPreConfigName))

		if etf2lPreErr == nil {
			fmt.Println("[Configs.Init] Etf2l's server pre-configs loaded!")
		} else {
			return "", etf2lPreErr
		}

		// ugc
	} else if c.League == "ugc" {
		if c.Type == LobbyTypeSixes {
			preConfigName = "/ugc/ugc_6v_base.cfg"

		} else if c.Type == LobbyTypeHighlander {
			preConfigName = "/ugc/ugc_HL_base.cfg"
		}
	}

	// pre config
	preConfig, preErr := ioutil.ReadFile(filepath.Clean(config.Constants.StaticFileLocation +
		ConfigsPath + preConfigName))

	if preErr == nil {
		fmt.Println("[Configs.Init] Server pre-configs loaded!")
	} else {
		return "", preErr
	}

	// get config file's data
	cfgData, cfgErr := ioutil.ReadFile(filepath.Clean(config.Constants.StaticFileLocation +
		ConfigsPath + "/" +
		strings.ToLower(c.League) + "/" +
		cfgName))

	if cfgErr == nil {
		fmt.Println("[Configs.Init] Server configs loaded!")
	} else {
		return "", cfgErr
	}

	var cfg string

	// insert etf2l pre config into server pre config
	if c.League == "etf2l" {
		cfg = string(etf2lPreConfig[:]) + string(preConfig[:]) + string(cfgData[:])
	} else {
		cfg = string(preConfig[:]) + string(cfgData[:])
	}

	return cfg, nil
}

func (c *ServerConfig) GetName() (string, error) {
	if c.League != "ugc" && c.League != "etf2l" {
		return "", errors.New("[Configs.GetName]: League can only be [ugc] or [etf2l]!")
	}

	if c.Type != LobbyTypeSixes && c.Type != LobbyTypeHighlander {
		return "", errors.New("[Configs.GetName]: The type you specified doesn't exists!")
	}

	// game type as string
	var t string

	if c.League == "etf2l" {
		switch {
		case c.Type == LobbyTypeSixes:
			t = "6v6"
		case c.Type == LobbyTypeHighlander:
			t = "9v9"
		}

	} else if c.League == "ugc" {
		switch {
		case c.Type == LobbyTypeSixes:
			t = "6v"
		case c.Type == LobbyTypeHighlander:
			t = "HL"
		}
	}

	// build config name
	// ugc -> 6v6 = ugc_6v_koth.cfg
	cfgName := strings.ToLower(c.League + "_" + t + "_" + c.Name + ".cfg")

	return cfgName, nil
}

func (c *ServerConfig) GetMapConfig(mapName string) (string, error) {
	fmt.Println("[Configs.GetMapConfig]: Getting config for map -> [" + mapName + "]")

	var mapConfig string

	// ugc 6v6
	if c.League == "ugc" && c.Type == LobbyTypeSixes && MapsData[mapName].V6 != nil {
		mapConfig = MapsData[mapName].V6.Ugc

		// ugc 9v9
	} else if c.League == "ugc" && c.Type == LobbyTypeHighlander && MapsData[mapName].V9 != nil {
		mapConfig = MapsData[mapName].V9.Ugc

		// etf2l 6v6
	} else if c.League == "etf2l" && c.Type == LobbyTypeSixes && MapsData[mapName].V6 != nil {
		mapConfig = MapsData[mapName].V6.Etf2l

		// etf2l 9v9
	} else if c.League == "etf2l" && c.Type == LobbyTypeHighlander && MapsData[mapName].V9 != nil {
		mapConfig = MapsData[mapName].V9.Etf2l

		// can't find any config in the maps.json
	} else {
		return "", errors.New("[Configs.GetMapConfig]: No config can be found for this map in this mode and league!")
	}

	return mapConfig, nil
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/op/go-logging"
)

type constants struct {
	GlobalChatRoom     string
	Port               string `env:"PORT"`
	Domain             string `env:"SERVER_DOMAIN"`
	OpenIDRealm        string `env:"SERVER_OPENID_REALM"`
	CookieDomain       string `env:"SERVER_COOKIE_DOMAIN"`
	LoginRedirectPath  string `env:"SERVER_REDIRECT_PATH"`
	CookieStoreSecret  string `env:"COOKIE_STORE_SECRET,secret"`
	StaticFileLocation string
	SessionName        string
	RPCAddr            string `env:"RPC_ADDR"`
	PaulingAddr        string `env:"PAULING_ADDR"`
	FumbleAddr         string `env:"FUMBLE_ADDR"`
	MumbleAddr         string `env:"MUMBLE_ADDR"`
	MumblePassword     string `env:"MUMBLE_PASSWORD"`
	SteamIDWhitelist   string `env:"STEAMID_WHITELIST"`
	ServerMockUp       bool   `env:"PAULING_DISABLE"`
	MockupAuth         bool   `env:"MOCKUP_AUTH"`
	GeoIP              string `env:"GEOIP_DB"`
	AllowedCorsOrigins []string

	// database
	DbAddr     string `env:"DATABASE_ADDR"`
	DbDatabase string `env:"DATABASE_NAME"`
	DbUsername string `env:"DATABASE_USERNAME"`
	DbPassword string `env:"DATABASE_PASSWORD,secret"`

	SteamDevApiKey string `env:"STEAM_API_KEY,secret"`
	SteamApiMockUp bool

	ProfilerEnable bool   `env:"PROFILER_ENABLE"`
	ProfilerPort   string `env:"PROFILER_ADDR"`

	SlackbotURL string `env:"SLACK_URL,secret"`
}

var Constants = constants{}

func SetupConstants() {
	setupDevelopmentConstants()
	if val := os.Getenv("DEPLOYMENT_ENV"); strings.ToLower(val) == "production" {
		setupProductionConstants()
	} else if val == "test" {
		setupTestConstants()
	} else if val == "travis_test" {
		setupTravisTestConstants()
	}

	ps := reflect.ValueOf(&Constants)

	for i := 0; i < ps.Elem().NumField(); i++ {
		field := ps.Elem().Field(i)
		opts := strings.Split(reflect.TypeOf(Constants).Field(i).Tag.Get("env"), ",")

		if len(opts) == 0 {
			continue
		}

		val := os.Getenv(opts[0])
		if val == "" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(val)
		case reflect.Bool:
			if val == "true" {
				field.SetBool(true)
			} else if val == "false" {
				field.SetBool(false)
			} else {
				logrus.Panicf("Invalid value for bool opt %s: %s", opts[0], val)
			}
		case reflect.Int:
			num, err := strconv.Atoi(val)
			if err != nil {
				logrus.Panicf("Invalid value for int opt %s: %s", opts[0], val)
			}
			field.SetInt(int64(num))
		}

		if len(opts) > 1 && opts[1] == "secret" {
			logrus.Debug("%s = %s", opts[0], logging.Redact(val))
			continue
		}

		logrus.Debug("%s = %s", opts[0], val)

	}

	if Constants.SteamDevApiKey == "your steam dev api key" && !Constants.SteamApiMockUp {
		logrus.Warning("Steam api key not provided, setting SteamApiMockUp to true")
		Constants.SteamApiMockUp = true
	}

}

func setupDevelopmentConstants() {
	Constants.GlobalChatRoom = "0"
	Constants.Port = "8080"
	Constants.RPCAddr = "localhost:8081"
	Constants.Domain = "http://localhost:8080"
	Constants.OpenIDRealm = "http://localhost:8080"
	Constants.CookieDomain = ""
	Constants.LoginRedirectPath = "http://localhost:8080/"
	Constants.CookieStoreSecret = ""
	Constants.SessionName = "defaultSession"
	Constants.StaticFileLocation = os.Getenv("GOPATH") + "/src/github.com/TF2Stadium/Helen/static"
	Constants.PaulingAddr = "localhost:8001"
	Constants.ServerMockUp = true
	Constants.AllowedCorsOrigins = []string{"*"}

	Constants.DbAddr = "127.0.0.1:5432"
	Constants.DbDatabase = "tf2stadium"
	Constants.DbUsername = "tf2stadium"
	Constants.DbPassword = "dickbutt" // change this

	Constants.SteamDevApiKey = "your steam dev api key"
	Constants.SteamApiMockUp = false

	Constants.ProfilerPort = "6060"
}

func setupProductionConstants() {
	// override production stuff here
	Constants.Port = "5555"
	Constants.CookieDomain = ".tf2stadium.com"
	Constants.ServerMockUp = false
}

func setupTestConstants() {
	Constants.DbAddr = "127.0.0.1:5432"
	Constants.DbDatabase = "TESTtf2stadium"
	Constants.DbUsername = "TESTtf2stadium"
	Constants.DbPassword = "dickbutt"

	Constants.ServerMockUp = true
	Constants.SteamApiMockUp = true
}

func setupTravisTestConstants() {
	Constants.DbAddr = "127.0.0.1:5432"
	Constants.DbDatabase = "tf2stadium"
	Constants.DbUsername = "postgres"
	Constants.DbPassword = ""

	Constants.ServerMockUp = true
	Constants.SteamApiMockUp = true
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package config

import (
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

var (
	GlobalChatRoom     string
	SteamApiMockUp     bool
	AllowedCorsOrigins []string
)

type constants struct {
	Address            string `envconfig:"SERVER_ADDR"` // -> HELEN_SERVER_DOMAIN
	OpenIDRealm        string `envconfig:"SERVER_OPENID_REALM"`
	CookieDomain       string `envconfig:"SERVER_COOKIE_DOMAIN"`
	LoginRedirectPath  string `envconfig:"SERVER_REDIRECT_PATH"`
	CookieStoreSecret  string `envconfig:"COOKIE_STORE_SECRET"`
	StaticFileLocation string
	SessionName        string
	RPCAddr            string `envconfig:"RPC_ADDR"`
	PaulingAddr        string `envconfig:"PAULING_ADDR"`
	FumbleAddr         string `envconfig:"FUMBLE_ADDR"`
	MumbleAddr         string `envconfig:"MUMBLE_ADDR"`
	MumblePassword     string `envconfig:"MUMBLE_PASSWORD"`
	SteamIDWhitelist   string `envconfig:"STEAMID_WHITELIST"`
	ServerMockUp       bool   `envconfig:"PAULING_DISABLE"`
	MockupAuth         bool   `envconfig:"MOCKUP_AUTH"`
	GeoIP              string `envconfig:"GEOIP_DB"`

	// database
	DbAddr     string `envconfig:"DATABASE_ADDR"`
	DbDatabase string `envconfig:"DATABASE_NAME"`
	DbUsername string `envconfig:"DATABASE_USERNAME"`
	DbPassword string `envconfig:"DATABASE_PASSWORD"`

	SteamDevApiKey string `envconfig:"STEAM_API_KEY"`

	ProfilerEnable bool   `envconfig:"PROFILER_ENABLE"`
	ProfilerPort   string `envconfig:"PROFILER_ADDR"`

	SlackbotURL string `envconfig:"SLACK_URL"`
}

var Constants = constants{}

func HTTPAddress() string {
	return "http://" + Constants.Address
}

func SetupConstants() {
	setupDevelopmentConstants()
	if val := os.Getenv("DEPLOYMENT_ENV"); strings.ToLower(val) == "production" {
		setupProductionConstants()
	} else if val == "test" {
		setupTestConstants()
	} else if val == "travis_test" {
		setupTravisTestConstants()
	}

	err := envconfig.Process("HELEN", &Constants)
	if err != nil {
		logrus.Fatal(err)
	}

	if Constants.SteamDevApiKey == "your steam dev api key" {
		logrus.Warning("Steam api key not provided, setting SteamApiMockUp to true")
	} else {
		SteamApiMockUp = false
	}

}

func setupDevelopmentConstants() {
	GlobalChatRoom = "0"
	Constants.RPCAddr = "localhost:8081"
	Constants.Address = "localhost:8080"
	Constants.OpenIDRealm = "http://localhost:8080"
	Constants.CookieDomain = ""
	Constants.LoginRedirectPath = "http://localhost:8080/"
	Constants.CookieStoreSecret = ""
	Constants.SessionName = "defaultSession"
	Constants.StaticFileLocation = os.Getenv("GOPATH") + "/src/github.com/TF2Stadium/Helen/static"
	Constants.PaulingAddr = "localhost:8001"
	Constants.ServerMockUp = true
	AllowedCorsOrigins = []string{"*"}

	Constants.DbAddr = "127.0.0.1:5432"
	Constants.DbDatabase = "tf2stadium"
	Constants.DbUsername = "tf2stadium"
	Constants.DbPassword = "dickbutt" // change this

	Constants.SteamDevApiKey = "your steam dev api key"
	SteamApiMockUp = true

	Constants.ProfilerPort = "6060"
}

func setupProductionConstants() {
	// override production stuff here
	Constants.CookieDomain = ".tf2stadium.com"
	Constants.ServerMockUp = false
}

func setupTestConstants() {
	Constants.DbAddr = "127.0.0.1:5432"
	Constants.DbDatabase = "TESTtf2stadium"
	Constants.DbUsername = "TESTtf2stadium"
	Constants.DbPassword = "dickbutt"

	Constants.ServerMockUp = true
	SteamApiMockUp = true
}

func setupTravisTestConstants() {
	Constants.DbAddr = "127.0.0.1:5432"
	Constants.DbDatabase = "tf2stadium"
	Constants.DbUsername = "postgres"
	Constants.DbPassword = ""

	Constants.ServerMockUp = true
	SteamApiMockUp = true
}

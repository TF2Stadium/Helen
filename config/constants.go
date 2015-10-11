// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package config

import (
	"github.com/TF2Stadium/Helen/helpers"
	"os"
	"strings"
)

type constants struct {
	GlobalChatRoom     string
	Port               string
	Domain             string
	OpenIDRealm        string
	CookieDomain       string
	LoginRedirectPath  string
	CookieStoreSecret  string
	StaticFileLocation string
	ChatLogsDir        string
	SessionName        string
	PaulingPort        string
	SocketMockUp       bool
	ServerMockUp       bool
	ChatLogsEnabled    bool
	MockupAuth         bool
	AllowedCorsOrigins []string

	// database
	DbHost     string
	DbPort     string
	DbDatabase string
	DbUsername string
	DbPassword string

	SteamDevApiKey string
	SteamApiMockUp bool
}

func overrideFromEnv(constant *string, name string) {
	val := os.Getenv(name)
	if "" != val {
		*constant = val
		helpers.Logger.Debug("%s set to %s", name, *constant)
	}
}

func overrideBoolFromEnv(constant *bool, name string) {
	val := os.Getenv(name)
	if val != "" {
		*constant = map[string]bool{
			"true":  true,
			"false": false,
		}[val]
		helpers.Logger.Debug("%s set to %t", name, *constant)
	}
}

var Constants constants

func SetupConstants() {
	Constants = constants{}

	setupDevelopmentConstants()
	if val := os.Getenv("DEPLOYMENT_ENV"); strings.ToLower(val) == "production" {
		setupProductionConstants()
	} else if val == "test" {
		setupTestConstants()
	} else if val == "travis_test" {
		setupTravisTestConstants()
	}

	overrideFromEnv(&Constants.Port, "PORT")
	overrideFromEnv(&Constants.ChatLogsDir, "CHAT_LOG_DIR")
	overrideFromEnv(&Constants.CookieStoreSecret, "COOKIE_STORE_SECRET")
	overrideFromEnv(&Constants.SteamDevApiKey, "STEAM_API_KEY")
	overrideFromEnv(&Constants.DbHost, "DATABASE_HOST")
	overrideFromEnv(&Constants.DbPort, "DATABASE_PORT")
	overrideFromEnv(&Constants.DbUsername, "DATABASE_USERNAME")
	overrideFromEnv(&Constants.DbPassword, "DATABASE_PASSWORD")
	overrideFromEnv(&Constants.PaulingPort, "PAULING_PORT")
	overrideFromEnv(&Constants.Domain, "SERVER_DOMAIN")
	overrideFromEnv(&Constants.OpenIDRealm, "SERVER_OPENID_REALM")
	overrideFromEnv(&Constants.CookieDomain, "SERVER_COOKIE_DOMAIN")
	overrideBoolFromEnv(&Constants.ChatLogsEnabled, "LOG_CHAT")
	overrideBoolFromEnv(&Constants.ServerMockUp, "PAULING_ENABLE")
	overrideBoolFromEnv(&Constants.MockupAuth, "MOCKUP_AUTH")
	overrideFromEnv(&Constants.LoginRedirectPath, "SERVER_REDIRECT_PATH")
	overrideFromEnv(&Constants.StaticFileLocation, "STATIC_FILE")
	// conditional assignments

	if Constants.SteamDevApiKey == "your steam dev api key" && !Constants.SteamApiMockUp {
		helpers.Logger.Warning("Steam api key not provided, setting SteamApiMockUp to true")
		Constants.SteamApiMockUp = true
	}

}

func setupDevelopmentConstants() {
	Constants.GlobalChatRoom = "0"
	Constants.Port = "8080"
	Constants.Domain = "http://localhost:8080"
	Constants.OpenIDRealm = "http://localhost:8080"
	Constants.CookieDomain = ""
	Constants.LoginRedirectPath = "http://localhost:8080/"
	Constants.CookieStoreSecret = "dev secret is very secret"
	Constants.SessionName = "defaultSession"
	Constants.StaticFileLocation = os.Getenv("GOPATH") + "/src/github.com/TF2Stadium/Helen/static/index.html"
	Constants.PaulingPort = "1234"
	Constants.ChatLogsDir = "."
	Constants.SocketMockUp = false
	Constants.ServerMockUp = true
	Constants.ChatLogsEnabled = false
	Constants.AllowedCorsOrigins = []string{"*"}

	Constants.DbHost = "127.0.0.1"
	Constants.DbPort = "5724"
	Constants.DbDatabase = "tf2stadium"
	Constants.DbUsername = "tf2stadium"
	Constants.DbPassword = "dickbutt" // change this

	Constants.SteamDevApiKey = "your steam dev api key"
	Constants.SteamApiMockUp = false
}

func setupProductionConstants() {
	// override production stuff here
	Constants.Port = "5555"
	Constants.ChatLogsDir = "."
	Constants.CookieDomain = ".tf2stadium.com"
	Constants.ServerMockUp = false
	Constants.ChatLogsEnabled = true
	Constants.SocketMockUp = false
}

func setupTestConstants() {
	Constants.DbHost = "127.0.0.1"
	Constants.DbDatabase = "TESTtf2stadium"
	Constants.DbUsername = "TESTtf2stadium"
	Constants.DbPassword = "dickbutt"

	Constants.ServerMockUp = true
	Constants.SteamApiMockUp = true
}

func setupTravisTestConstants() {
	Constants.DbHost = "127.0.0.1"
	Constants.DbDatabase = "tf2stadium"
	Constants.DbUsername = "postgres"
	Constants.DbPassword = ""

	Constants.ServerMockUp = true
	Constants.SteamApiMockUp = true
}

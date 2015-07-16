package config

import (
	"os"
	"strings"
)

type constants struct {
	Port               string
	Domain             string
	OpenIDRealm        string
	LoginRedirectPath  string
	CookieStoreSecret  string
	StaticFileLocation string
	SessionName        string
	SocketMockUp       bool
	AllowedCorsOrigins []string

	// database
	DbHosts        string
	DbDatabase     string
	DbTestDatabase string
	DbUsername     string
	DbPassword     string

	DbLobbiesCollection string
	DbPlayersCollection string
}

func overrideFromEnv(constant *string, name string) {
	val := os.Getenv(name)
	if "" != val {
		*constant = val
	}
}

var Constants constants

func SetupConstants() {
	if val := os.Getenv("DEPLOYMENT_ENV"); strings.ToLower(val) != "production" {
		setupDevelopmentConstants()
	} else {
		setupProductionConstants()
	}

	overrideFromEnv(&Constants.Port, "PORT")
	overrideFromEnv(&Constants.CookieStoreSecret, "COOKIE_STORE_SECRET")

	// TODO: database url from env
	// TODO: database info from env
}

func setupDevelopmentConstants() {
	Constants = constants{
		Port:               "8080",
		Domain:             "http://localhost:8080",
		OpenIDRealm:        "http://localhost:8080",
		LoginRedirectPath:  "http://localhost:8080/",
		CookieStoreSecret:  "dev secret is very secret",
		SessionName:        "defaultSession",
		StaticFileLocation: os.Getenv("GOPATH") + "/src/github.com/TeamPlayTF/Server/static",
		SocketMockUp:       false,
		AllowedCorsOrigins: []string{"*"},

		DbHosts:        "127.0.0.1:27017",
		DbDatabase:     "teamplaytf",
		DbTestDatabase: "TESTteamplaytf",
		DbUsername:     "teamplaytf",
		DbPassword:     "dickbutt", // change this

		DbLobbiesCollection: "lobbies", // change this
		DbPlayersCollection: "players", // change this
	}
}

func setupProductionConstants() {
	// TODO
	Constants = constants{
		Port:               "5555",
		Domain:             "http://localhost:8080",
		OpenIDRealm:        "http://localhost:8080",
		CookieStoreSecret:  "dev secret is very secret",
		StaticFileLocation: os.Getenv("GOPATH") + "/src/github.com/TeamPlayTF/Server/static",
		SocketMockUp:       false,
		AllowedCorsOrigins: []string{"http://teamplay.tf", "http://api.teamplay.tf"},

		DbHosts:        "127.0.0.1:27017",
		DbDatabase:     "teamplaytf",
		DbTestDatabase: "TESTteamplaytf",
		DbUsername:     "teamplaytf",
		DbPassword:     "dickbutt", // change this
	}
}

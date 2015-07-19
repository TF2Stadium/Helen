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
	DbHost         string
	DbPort         string
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
		StaticFileLocation: os.Getenv("GOPATH") + "/src/github.com/TF2Stadium/Server/static",
		SocketMockUp:       false,
		AllowedCorsOrigins: []string{"*"},

		DbHost:         "127.0.0.1",
		DbPort:         "5724",
		DbDatabase:     "tf2stadium",
		DbTestDatabase: "TESTtf2stadium",
		DbUsername:     "tf2stadium",
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
		StaticFileLocation: os.Getenv("GOPATH") + "/src/github.com/TF2Stadium/Server/static",
		SocketMockUp:       false,
		AllowedCorsOrigins: []string{"http://tf2stadium.com", "http://api.tf2stadium.com"},

		DbHost:         "127.0.0.1",
		DbPort:         "5724",
		DbDatabase:     "tf2stadium",
		DbTestDatabase: "TESTtf2stadium",
		DbUsername:     "tf2stadium",
		DbPassword:     "dickbutt", // change this
	}
}

package config

import (
	"os"
	"strings"
)

type constants struct {
	Port               string
	Domain             string
	OpenIDRealm        string
	CookieStoreSecret  string
	StaticFileLocation string
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
}

func setupDevelopmentConstants() {
	Constants = constants{
		Port:               "8080",
		Domain:             "http://localhost:8080",
		OpenIDRealm:        "http://localhost:8080",
		CookieStoreSecret:  "dev secret is very secret",
		StaticFileLocation: os.Getenv("GOPATH") + "/src/github.com/TeamPlayTF/Server/static",
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
	}
}

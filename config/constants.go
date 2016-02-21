// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package config

import (
	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
)

var (
	GlobalChatRoom     string = "0"
	AllowedCorsOrigins        = []string{"*"}
)

type constants struct {
	ListenAddress      string `envconfig:"SERVER_ADDR" default:"localhost:8080"`
	PublicAddress      string `envconfig:"PUBLIC_ADDR"` // should include schema
	OpenIDRealm        string `envconfig:"SERVER_OPENID_REALM" default:"http://localhost:8080"`
	CookieDomain       string `envconfig:"SERVER_COOKIE_DOMAIN" default:""`
	LoginRedirectPath  string `envconfig:"SERVER_REDIRECT_PATH" default:"http://localhost:8080/"`
	CookieStoreSecret  string `envconfig:"COOKIE_STORE_SECRET" default:"secret"`
	StaticFileLocation string
	SessionName        string `envconfig:"COOKIE_SESSION_NAME" default:"defaultSession"`
	//RPCAddr            string `envconfig:"RPC_ADDR" default:"localhost:8081"`
	MumbleAddr       string `envconfig:"MUMBLE_ADDR"`
	MumblePassword   string `envconfig:"MUMBLE_PASSWORD"`
	SteamIDWhitelist string `envconfig:"STEAMID_WHITELIST"`
	MockupAuth       bool   `envconfig:"MOCKUP_AUTH" default:"false"`
	GeoIP            bool   `envconfig:"GEOIP" default:"false"`
	ServeStatic      bool   `envconfig:"SERVE_STATIC" default:"true"`
	RabbitMQURL      string `envconfig:"RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/"`
	PaulingQueue     string `envconfig:"PAULING_QUEUE" default:"pauling"`
	FumbleQueue      string `envconfig:"FUMBLE_QUEUE" default:"fumble"`
	RabbitMQQueue    string `envconfig:"RABBITMQ_QUEUE" default:"events"`
	RabbitMQExchange string `envconfig:"RABBITMQ_EXCHANGE" default:"helen-fanout"`

	// database
	DbAddr     string `envconfig:"DATABASE_ADDR" default:"127.0.0.1:5432"`
	DbDatabase string `envconfig:"DATABASE_NAME" default:"tf2stadium"`
	DbUsername string `envconfig:"DATABASE_USERNAME" default:"tf2stadium"`
	DbPassword string `envconfig:"DATABASE_PASSWORD" default:"dickbutt"`

	SteamDevAPIKey string `envconfig:"STEAM_API_KEY"`

	ProfilerAddr string `envconfig:"PROFILER_ADDR"`

	SlackbotURL        string `envconfig:"SLACK_URL"`
	TwitchClientID     string `envconfig:"TWITCH_CLIENT_ID"`
	TwitchClientSecret string `envconfig:"TWITCH_CLIENT_SECRET"`

	// EtcdAddr    string `envconfig:"ETCD_ADDR"`
	// EtcdService string `envconfig:"ETCD_SERVICE"`
}

var Constants = constants{}

func SetupConstants() {
	err := envconfig.Process("HELEN", &Constants)
	if err != nil {
		logrus.Fatal(err)
	}

	if Constants.SteamDevAPIKey == "" {
		logrus.Warning("Steam api key not provided, setting SteamApiMockUp to true")
	}

	if Constants.PublicAddress == "" {
		Constants.PublicAddress = "http://" + Constants.ListenAddress
	}
}

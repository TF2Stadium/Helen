// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package config

import (
	"net/url"
	"os"
	"reflect"
	"text/template"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var (
	mdTableTemplate = template.Must(template.New("doc").Parse(`
| Environment Variable | Description |
|----------------------|-------------|{{range .}}
|    ` + "`{{index . 0}}`" + `     |{{index . 1}}|{{end}}
`))
)

type constants struct {
	ListenAddress     string   `envconfig:"SERVER_ADDR" default:"localhost:8080" doc:"Address to serve on."`
	PublicAddress     string   `envconfig:"PUBLIC_ADDR" doc:"Publicly accessible address for the server, requires schema"`
	OpenIDRealm       string   `envconfig:"SERVER_OPENID_REALM" default:"http://localhost:8080" doc:"The OpenID Realm (See: [Section 9.2 of the OpenID Spec](https://openid.net/specs/openid-authentication-2_0-12.html#realms))"`
	AllowedOrigins    []string `envconfig:"ALLOWED_ORIGINS" default:"*"`
	CookieDomain      string   `envconfig:"SERVER_COOKIE_DOMAIN" default:"" doc:"Cookie URL domain"`
	LoginRedirectPath string   `envconfig:"SERVER_REDIRECT_PATH" default:"http://localhost:8080/" doc:"URL to redirect user to after a successful login"`
	CookieStoreSecret string   `envconfig:"COOKIE_STORE_SECRET" default:"secret" doc:"base64 encoded key to use for encrypting cookies"`
	MumbleAddr        string   `envconfig:"MUMBLE_ADDR" doc:"Mumble Address"`
	SteamIDWhitelist  string   `envconfig:"STEAMID_WHITELIST" doc:"SteamID Group XML page to use to filter logins"`
	MockupAuth        bool     `envconfig:"MOCKUP_AUTH" default:"false" doc:"Enable Mockup Authentication"`
	GeoIP             bool     `envconfig:"GEOIP" default:"false" doc:"Enable geoip support for getting the location of game servers"`
	ServeStatic       bool     `envconfig:"SERVE_STATIC" default:"true" doc:"Serve /static/"`
	RabbitMQURL       string   `envconfig:"RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/" doc:"URL for AMQP server"`
	PaulingQueue      string   `envconfig:"PAULING_QUEUE" default:"pauling" doc:"Name of queue over which RPC calls to Pauling are sent"`
	TwitchBotQueue    string   `envconfig:"TWITCHBOT_QUEUE" default:"twitchbot" doc:"Name of queue over which RPC calls to Pauling are sent"`
	FumbleQueue       string   `envconfig:"FUMBLE_QUEUE" default:"fumble" doc:"Name of queue over which RPC calls to Fumble are sent"`
	RabbitMQQueue     string   `envconfig:"RABBITMQ_QUEUE" default:"events" doc:"Name of queue over which events are sent"`

	// database
	DbAddr     string `envconfig:"DATABASE_ADDR" default:"127.0.0.1:5432" doc:"Database Address"`
	DbDatabase string `envconfig:"DATABASE_NAME" default:"tf2stadium" doc:"Database Name"`
	DbUsername string `envconfig:"DATABASE_USERNAME" default:"tf2stadium" doc:"Database username"`
	DbPassword string `envconfig:"DATABASE_PASSWORD" default:"dickbutt" doc:"Database password"`

	SteamDevAPIKey string `envconfig:"STEAM_API_KEY" doc:"Steam API Key"`

	ProfilerAddr string `envconfig:"PROFILER_ADDR" doc:"Address to serve the web-based profiler over"`

	SlackbotURL        string        `envconfig:"SLACK_URL" doc:"Slack webhook URL"`
	SentryDSN          string        `envconfig:"SENTRY_DSN" doc:"Sentry DSN"`
	DiscordToken       string        `envconfig:"DISCORD_TOKEN" doc:"Discord Token"`
	DiscordGuildId     string        `envconfig:"DISCORD_GUILD_ID" doc:"Discord Guild ID"`
	Environment        string        `envconfig:"DEPLOYED_ENV" default:"development" doc:"Deployment environment"`
	TwitchClientID     string        `envconfig:"TWITCH_CLIENT_ID" doc:"Twitch API Client ID"`
	TwitchClientSecret string        `envconfig:"TWITCH_CLIENT_SECRET" doc:"Twitch API Client Secret"`
	ServemeAPIKey      string        `envconfig:"SERVEME_API_KEY" doc:"serveme.tf API Key"`
	HealthChecks       bool          `envconfig:"HEALTH_CHECKS" default:"false" doc:"Enable health checks"`
	SecureCookies      bool          `envconfig:"SECURE_COOKIE" doc:"Enable 'secure' flag on cookies" default:"false"`
	FilteredWords      []string      `envconfig:"FILTERED_WORDS"`
	ChatRateLimit      time.Duration `envconfig:"CHAT_RATE_LIMIT"`
	DemosFolder        string        `envconfig:"DEMOS_FOLDER" doc:"Folder to store STV demos in" default:"demos"`
}

var Constants = constants{}

func init() {
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
	if Constants.MockupAuth {
		logrus.Warning("Mockup authentication enabled.")
	}

	_, err = url.Parse(Constants.PublicAddress)
	if err != nil {
		logrus.Fatal("Couldn't parse HELEN_PUBLIC_ADDR - ", err)
	}
	_, err = url.Parse(Constants.LoginRedirectPath)
	if err != nil {
		logrus.Fatal("Couldn't parse HELEN_SERVER_REDIRECT_PATH - ", err)
	}

	if Constants.GeoIP {
		logrus.Info("GeoIP support enabled")
	}
}

func PrintConfigDoc() {
	var data [][]string
	t := reflect.TypeOf(constants{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("envconfig") == "" {
			continue
		}
		data = append(data, []string{field.Tag.Get("envconfig"), field.Tag.Get("doc")})
	}

	mdTableTemplate.Execute(os.Stdout, data)
}

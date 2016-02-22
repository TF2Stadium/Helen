Helen
=====
Helen is the backend server component for the tf2stadium.com project written in Go.

[![Build Status](https://circleci.com/gh/TF2Stadium/Helen/tree/dev.svg?style=svg)](https://circleci.com/gh/TF2Stadium/Helen/tree/dev)
[![Go Report Card](https://img.shields.io/badge/go_report-A-brightgreen.svg)](http://goreportcard.com/report/TF2Stadium/Helen)
[![GitHub license](https://img.shields.io/badge/license-GPLv3-blue.svg?style=flat-square)](https://raw.githubusercontent.com/TF2Stadium/Helen/master/COPYING)
[![Stories in Ready](https://badge.waffle.io/TF2Stadium/Helen.png?label=ready&title=Ready)](http://waffle.io/TF2Stadium/Helen)

[Coverage Status](https://tf2stadium.github.io/coverage/)

### Setup

The project uses PostgreSQL (PSQL) as a database. Default development account data can be found at [database/setup.md](../master/database/setup.md).

Running this project requires configuring it via environment
variables.

| Environment Variable | Description |
|----------------------|-------------|
|    `SERVER_ADDR`     |Address to serve on.|
|    `PUBLIC_ADDR`     |Publicly accessible address for the server, requires schema|
|    `SERVER_OPENID_REALM`     |The OpenID Realm (See: [Section 9.2 of the OpenID Spec](https://openid.net/specs/openid-authentication-2_0-12.html#realms))|
|    `SERVER_COOKIE_DOMAIN`     |Cookie URL domain|
|    `SERVER_REDIRECT_PATH`     |URL to redirect user to after a successful login|
|    `COOKIE_STORE_SECRET`     |base64 encoded key to use for encrypting cookies|
|    `MUMBLE_ADDR`     |Mumble Address|
|    `MUMBLE_PASSWORD`     |Mumble Password|
|    `STEAMID_WHITELIST`     |SteamID Group XML page to use to filter logins|
|    `MOCKUP_AUTH`     |Enable Mockup Authentication|
|    `GEOIP`     |Enable geoip support for getting the location of game servers|
|    `SERVE_STATIC`     |Serve /static/|
|    `RABBITMQ_URL`     |URL for AMQP server|
|    `PAULING_QUEUE`     |Name of queue over which RPC calls to Pauling are sent|
|    `FUMBLE_QUEUE`     |Name of queue over which RPC calls to Fumble are sent|
|    `RABBITMQ_QUEUE`     |Name of queue over which events are sent|
|    `RABBITMQ_EXCHANGE`     |Name of queue over which socket messages are fanned out to other Helen instances|
|    `DATABASE_ADDR`     |Database Address|
|    `DATABASE_NAME`     |Database Name|
|    `DATABASE_USERNAME`     |Database username|
|    `DATABASE_PASSWORD`     |Database password|
|    `STEAM_API_KEY`     |Steam API Key|
|    `PROFILER_ADDR`     |Address to serve the web-based profiler over|
|    `SLACK_URL`     |Slack webhook URL|
|    `TWITCH_CLIENT_ID`     |Twitch API Client ID|
|    `TWITCH_CLIENT_SECRET`     |Twitch API Client Secret|

### Structure
The code is divided into multiple packages that follow the usual web application structure:
* models go in `models`
* controllers go in `controllers`
* database go in `database`, migration code goes to [database/migrations](../master/database/migrations)
* routes go in `routes/routes.go`
* helpers go in `helpers`

### Contributing
1. Fork this repository - http://github.com/TF2Stadium/Helen/fork
2. Create your feature branch - `git checkout -b my-new-feature`
3. Commit your changes - `git commit`
4. Push - `git push origin my-new-feature`
5. Create a Pull Request.

**Before creating a Pull Request:**

1. Ensure the code matches the Go style guidelines mentioned [Here](https://github.com/golang/go/wiki/CodeReviewComments). Code can be formatted with the `go fmt` tool.
2. Ensure existing tests pass (with `go test ./...`), or are updated appropriately.
3. For new features, you should add new tests.
4. The pull request should be squashed (no more than 1 temporary commit per 100 loc, more info [here](http://eli.thegreenplace.net/2014/02/19/squashing-github-pull-requests-into-a-single-commit))

### License

Helen is licensed under the GNU Public License v3.

This product includes GeoLite2 data created by MaxMind, available from [http://www.maxmind.com"](http://www.maxmind.com)

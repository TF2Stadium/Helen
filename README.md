Helen - DEV BRANCH
==================
Helen is the backend server component for the tf2stadium.com project written in Go.

[![Build Status](https://circleci.com/gh/TF2Stadium/Helen/tree/dev.svg?style=svg)](https://circleci.com/gh/TF2Stadium/Helen/tree/dev)
[![Go Report Card](https://img.shields.io/badge/Go%20Report%20Card-score-blue.svg?style=flat-square)](http://goreportcard.com/report/TF2Stadium/Helen)
[![GitHub license](https://img.shields.io/badge/license-GPLv3-blue.svg?style=flat-square)](https://raw.githubusercontent.com/TF2Stadium/Helen/master/COPYING)
[![Stories in Ready](https://badge.waffle.io/TF2Stadium/Helen.png?label=ready&title=Ready)](http://waffle.io/TF2Stadium/Helen)

[Coverage Status](https://tf2stadium.github.io/coverage/)

### Setup

The project uses PostgreSQL (PSQL) as a database. Default development account data can be found at [database/setup.md](../master/database/setup.md).

Running this project requires configuring it via environment
variables.

| Env Var Name                 | Description                                                                                                                                      |
|------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| `HELEN_SERVER_ADDR`          | Address to listen on, eg `localhost:8080`                                                                                                        |
| `HELEN_PUBLIC_ADDR`          | Publicly accessible address for the server, requires schema, eg `http://example.com`. If not supplied, will default to `http://` + `SERVER_ADDR` |
| `HELEN_SERVER_OPENID_REALM`  | The OpenID Realm (See: [Section 9.2 of the OpenID Spec](https://openid.net/specs/openid-authentication-2_0-12.html#realms))                      |
| `HELEN_SERVER_COOKIE_DOMAIN` | Cookie URL domain                                                                                                                                |
| `HELEN_RPC_ADDR`             | Address to listen on for RPC requests                                                                                                            |
| `HELEN_PAULING_ADDR`         | Address to connect to [Pauling](github.com/TF2Stadium/Pauling) on.                                                                               |
| `HELEN_FUMBLE_ADDR`          | Address to connect to [Fumble](github.com/TF2Stadium/Fumble) on.                                                                                 |
| `HELEN_MUMBLE_ADDR`          | Mumble server address for lobbies.                                                                                                               |
| `HELEN_MUMBLE_PASSWORD`      | Mumble server password for lobbies.                                                                                                              |
| `HELEN_STEAMID_WHITELIST`    | If set, only players in the steamgroup can login to the site.                                                                                    |
| `HELEN_PAULING_DISABLE`      | Disable Pauling support by setting this to "false".                                                                                              |
| `HELEN_MOCKUP_AUTH`          | Allow Mock logins for testing.                                                                                                                   |
| `HELEN_GEOIP_DB`             | Path to GeoIP Database to use for geolocating lobbies.                                                                                           |
| `HELEN_DATABASE_ADDR`        | PSQL Address                                                                                                                                     |
| `HELEN_DATABASE_NAME`        | PSQL Database name                                                                                                                               |
| `HELEN_DATABASE_USERNAME`    | PSQL Username                                                                                                                                    |
| `HELEN_DATABASE_PASSWORD`    | PSQL Password                                                                                                                                    |
| `HELEN_STEAM_API_KEY`        | Steam API key, for steam integration                                                                                                             |
| `HELEN_PROFILER_ENABLE`      | Enable profiler if set to "true".                                                                                                                |
| `HELEN_PROFILER_ADDR`        | Address for profiler to listen on.                                                                                                               |
| `HELEN_SLACK_URL`            | Slack URL for slack integration.                                                                                                                 |

### Structure
The code is divided into multiple packages that follow the usual web application structure:
* models go in `models`
* controllers go in `controllers`
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

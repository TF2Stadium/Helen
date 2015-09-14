# TF2Stadium
TF2Stadium Server is the backend server component for the tf2stadium.com project written in Go.

[![Build Status](https://drone.io/github.com/TF2Stadium/Helen/status.png)](https://drone.io/github.com/TF2Stadium/Helen/latest)

### Setup
The project uses postgres as a database. Default development account data can be found at  [database/setup.md](../blob/master/database/setup.md).

### Structure
The code is divided into multiple packages that follow the usual web application structure:
* models go in `models`
* controllers go in `controllers`
* routes go in `routes/routes.go`
* TODO views currently go to static, until work on frontend code starts

### Contribution guidelines
The project uses the Pull Request workflow to contribute code. More info on that here: https://help.github.com/articles/using-pull-requests/.

**Each pull request must pass all existing tests (go test ./...) and include new appropriate tests.**

The pull request should be squashed (no more than 1 temporary commit per 100 loc, more info here: http://eli.thegreenplace.net/2014/02/19/squashing-github-pull-requests-into-a-single-commit)

Submitted code has to be formatted with `go fmt`.

Helen
=====
Helen is the backend server component for the tf2stadium.com project written in Go.

[![Build Status](https://drone.io/github.com/TF2Stadium/Helen/status.png)](https://drone.io/github.com/TF2Stadium/Helen/latest)
[![Go Report Card](https://img.shields.io/badge/Go%20Report%20Card-score-blue.svg?style=flat-square)](http://goreportcard.com/report/TF2Stadium/Helen)
[![GitHub license](https://img.shields.io/badge/license-GPLv3-blue.svg?style=flat-square)](https://raw.githubusercontent.com/TF2Stadium/Helen/master/COPYING)

### Setup
The project uses postgres as a database. Default development account data can be found at  [database/setup.md](../master/database/setup.md).

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

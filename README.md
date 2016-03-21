Helen
=====
Helen is the backend server component for the tf2stadium.com project written in Go.

[![Build Status](https://travis-ci.org/TF2Stadium/Helen.svg?branch=master)](https://travis-ci.org/TF2Stadium/Helen)
[![](https://badge.imagelayers.io/tf2stadium/helen:latest.svg)](https://imagelayers.io/?images=tf2stadium/helen:latest 'Get your own badge on imagelayers.io')
[![Go Report Card](https://img.shields.io/badge/go_report-A-brightgreen.svg)](http://goreportcard.com/report/TF2Stadium/Helen)
[![GitHub license](https://img.shields.io/badge/license-GPLv3-blue.svg?style=flat-square)](https://raw.githubusercontent.com/TF2Stadium/Helen/master/COPYING)
[![Stories in Ready](https://badge.waffle.io/TF2Stadium/Helen.png?label=ready&title=Ready)](http://waffle.io/TF2Stadium/Helen)

[Coverage Status](https://tf2stadium.github.io/coverage/)

### Requirements

* Go >= 1.5
* PostgreSQL (with the `hstore` extension installed) Default development account data can be found at [database/setup.md](../master/database/setup.md)
* RabbitMQ
* [go-bindata](https://github.com/jteeuwen/go-bindata)

### Installation

Running this project requires configuring it via environment variables, documentation for which can be found on [CONFIG.md](./master/CONFIG.md)

1. `go get github.com/TF2Stadium/Helen`
2. `cd $(GOPATH)/src/github.com/TF2Stadium/Helen`
3. `make assets -B`
4. `make static`

To build docker images, use `docker build -t tf2stadium/helen .`

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

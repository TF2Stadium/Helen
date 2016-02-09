// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/vibhavp/rpcconn"
)

var (
	pauling *rpcconn.Client
	fumble  *rpcconn.Client
)

func ConnectRPC() {
	var err error
	var addr string

	if config.Constants.PaulingAddr != "" {
		addr = config.Constants.PaulingAddr

		if strings.HasPrefix(addr, "etcd:") {
			addr, err = helpers.GetAddr(strings.Split(addr, ":")[1])
			if err != nil {
				logrus.Fatal(err)
			}
		}

		pauling, err = rpcconn.DialHTTP("tcp", addr)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	if config.Constants.FumbleAddr != "" {
		addr = config.Constants.FumbleAddr

		if strings.HasPrefix(addr, "etcd:") {
			addr, err = helpers.GetAddr(strings.Split(addr, ":")[1])
			if err != nil {
				logrus.Fatal(err)
			}
		}

		fumble, err = rpcconn.DialHTTP("tcp", config.Constants.FumbleAddr)
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

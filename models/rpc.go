// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/etcd"
	"github.com/vibhavp/rpcconn"
)

var (
	pauling *rpcconn.Client
	fumble  *rpcconn.Client
)

func ConnectRPC() {
	var err error

	if config.Constants.PaulingAddr != "" {
		pauling, err = rpcconn.DialHTTP("tcp", etcd.Address{Address: config.Constants.PaulingAddr})
		if err != nil {
			logrus.Fatal(err)
		}

		pauling.Call("Pauling.Ping", struct{}{}, &struct{}{})
	}

	if config.Constants.FumbleAddr != "" {
		fumble, err = rpcconn.DialHTTP("tcp", etcd.Address{Address: config.Constants.FumbleAddr})
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"flag"
	"net/rpc"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/vibhavp/amqp-rpc"
)

var (
	pauling         *rpc.Client
	fumble          *rpc.Client
	paulingDisabled = flag.Bool("disable_pauling", true, "disable pauling")
	fumbleDisabled  = flag.Bool("disable_fumble", true, "disable fumble")
)

func ConnectRPC() {
	codec, err := amqprpc.NewClientCodec(helpers.AMQPConn, config.Constants.PaulingQueue, amqprpc.JSONCodec{})
	if err != nil {
		logrus.Fatal(err)
	}

	pauling = rpc.NewClientWithCodec(codec)

	codec, err = amqprpc.NewClientCodec(helpers.AMQPConn, config.Constants.FumbleQueue, amqprpc.JSONCodec{})
	fumble = rpc.NewClientWithCodec(codec)
	if err != nil {
		logrus.Fatal(err)
	}

}

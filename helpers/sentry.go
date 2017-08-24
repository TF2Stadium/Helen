// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"github.com/sirupsen/logrus"
  "github.com/getsentry/raven-go"
	"github.com/TF2Stadium/Helen/config"
)

var Raven *raven.Client = nil

func init() {
	dsn := config.Constants.SentryDSN
	if dsn != "" {
		var err error
		Raven, err = raven.New(dsn)
		Raven.SetEnvironment(config.Constants.Environment)

		if err != nil {
			logrus.Fatal(err)
		}
	}
}

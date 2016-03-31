// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package database

import (
	"os"
	"strconv"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	_ "github.com/TF2Stadium/Helen/helpers"
	"github.com/stretchr/testify/assert"
)

var steamid = "76561198074578368"

func TestDatabasePing(t *testing.T) {
	ci := os.Getenv("CI")
	if ci == "true" {
		config.Constants.DbUsername = "postgres"
		config.Constants.DbDatabase = "travis_ci_test"
		config.Constants.DbPassword = ""
	}

	logrus.Debug("[Test.Database] IsTest? " + strconv.FormatBool(IsTest))
	Init()
	assert.Nil(t, DB.DB().Ping())
}

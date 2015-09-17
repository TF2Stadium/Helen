// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package database

import (
	"os"
	"strconv"
	"testing"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/stretchr/testify/assert"
)

var steamid = "76561198074578368"

func init() {
	helpers.InitLogger()
}

func TestDatabasePing(t *testing.T) {
	if os.Getenv("DEPLOYMENT_ENV") == "" {
		os.Setenv("DEPLOYMENT_ENV", "test")
		defer os.Unsetenv("DEPLOYMENT_ENV")
	}
	config.SetupConstants()

	helpers.Logger.Debug("[Test.Database] IsTest? " + strconv.FormatBool(IsTest))
	Init()
	assert.Nil(t, DB.DB().Ping())
}

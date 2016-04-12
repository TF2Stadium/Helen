// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"os"
	"sync"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	_ "github.com/TF2Stadium/Helen/helpers/authority"
)

var cleaningMutex sync.Mutex

var o = new(sync.Once)

func CleanupDB() {
	cleaningMutex.Lock()
	defer cleaningMutex.Unlock()

	o.Do(func() {
		ci := os.Getenv("CI")
		config.Constants.DbAddr = "127.0.0.1:5432"
		config.Constants.SteamDevAPIKey = ""

		if ci == "true" {
			config.Constants.DbUsername = "postgres"
			config.Constants.DbDatabase = "travis_ci_test"
			config.Constants.DbPassword = ""
		} else {
			config.Constants.DbDatabase = "TESTtf2stadium"
			config.Constants.DbUsername = "TESTtf2stadium"
			config.Constants.DbPassword = "dickbutt"
		}

		database.Init()
	})

	database.DB.Exec("drop schema public cascade")
	database.DB.Exec("create schema public")
	migrations.Do()
}

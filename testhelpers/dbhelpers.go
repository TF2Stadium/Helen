// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"os"
	"sync"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers/authority"
)

var cleaningMutex sync.Mutex

func CleanupDB() {
	cleaningMutex.Lock()
	defer cleaningMutex.Unlock()
	os.Setenv("DEPLOYMENT_ENV", "test")
	config.SetupConstants()
	database.Init()
	authority.RegisterTypes()
	database.DB.Exec("DROP SCHEMA public CASCADE;")
	database.DB.Exec("CREATE SCHEMA public;")

	migrations.Do()

	stores.SetupStores()
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	//	"fmt"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"os"
	"sync"
)

var cleaningMutex sync.Mutex

func CleanupDB() {
	cleaningMutex.Lock()
	defer cleaningMutex.Unlock()
	if os.Getenv("DEPLOYMENT_ENV") == "" {
		os.Setenv("DEPLOYMENT_ENV", "test")
		defer os.Unsetenv("DEPLOYMENT_ENV")
	}
	config.SetupConstants()
	database.Init()

	database.DB.Exec("DROP SCHEMA public CASCADE;")
	database.DB.Exec("CREATE SCHEMA public;")

	migrations.Do()

	stores.SetupStores()
}

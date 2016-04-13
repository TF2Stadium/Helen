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
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/models/chat"
	"github.com/TF2Stadium/Helen/models/gameserver"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/player"
)

var (
	cleaningMutex sync.Mutex
	o             = new(sync.Once)
	tables        = []interface{}{
		&player.Player{},
		&lobby.Lobby{},
		&lobby.LobbySlot{},
		&gameserver.ServerRecord{},
		&player.PlayerStats{},
		&models.AdminLogEntry{},
		&player.PlayerBan{},
		&chat.ChatMessage{},
		&lobby.Requirement{},
		&gameserver.StoredServer{},
		&player.Report{},
	}
)

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
		for _, table := range tables {
			database.DB.DropTable(table)
		}
		migrations.Do()
	})

}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package testhelpers

import (
	"sync"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers/authority"
)

var cleaningMutex sync.Mutex

var o = new(sync.Once)

func CleanupDB() {
	cleaningMutex.Lock()
	defer cleaningMutex.Unlock()

	o.Do(func() {
		config.SetupConstants()

		config.Constants.DbAddr = "127.0.0.1:5432"
		config.Constants.DbDatabase = "TESTtf2stadium"
		config.Constants.DbUsername = "TESTtf2stadium"
		config.Constants.DbPassword = "dickbutt"
		config.SteamApiMockUp = true

		database.Init()
		authority.RegisterTypes()
		migrations.Do()
		stores.SetupStores()
	})

	tables := []string{
		"admin_log_entries",
		"banned_players_lobbies",
		"chat_messages",
		"http_sessions",
		"lobbies",
		"lobby_slots",
		"player_bans",
		"player_stats",
		"players",
		"server_records",
		"spectators_players_lobbies",
		"requirements",
	}
	for _, table := range tables {
		database.DB.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY")
	}

}

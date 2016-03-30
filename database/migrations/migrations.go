// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"bytes"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/assets"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/models/chat"
	"github.com/TF2Stadium/Helen/models/gameserver"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/gchaincl/dotsql"
)

var once = new(sync.Once)

func Do() {
	database.DB.Exec("CREATE EXTENSION IF NOT EXISTS hstore")
	database.DB.AutoMigrate(&player.Player{})
	database.DB.AutoMigrate(&lobby.Lobby{})
	database.DB.AutoMigrate(&lobby.LobbySlot{})
	database.DB.AutoMigrate(&gameserver.Server{})
	database.DB.AutoMigrate(&player.Stats{})
	database.DB.AutoMigrate(&models.AdminLogEntry{})
	database.DB.AutoMigrate(&player.Ban{})

	database.DB.Model(&lobby.LobbySlot{}).AddUniqueIndex("idx_lobby_slot_lobby_id_slot", "lobby_id", "slot")
	database.DB.Model(&lobby.LobbySlot{}).AddUniqueIndex("idx_lobby_id_player_id", "lobby_id", "player_id")
	database.DB.AutoMigrate(&chat.ChatMessage{})
	database.DB.AutoMigrate(&lobby.Requirement{})
	database.DB.AutoMigrate(&Constant{})
	database.DB.AutoMigrate(&gameserver.StoredServer{})

	once.Do(func() {
		checkSchema()
		dot, err := dotsql.Load(bytes.NewBuffer(assets.MustAsset("assets/views.sql")))
		if err != nil {
			logrus.Fatal(err)
		}

		_, err = dot.Exec(database.DB.DB(), "create-player-slots-view")
		if err != nil {
			logrus.Fatal(err)
		}
	})
}

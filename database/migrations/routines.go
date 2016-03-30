// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"math/rand"
	"strconv"

	"github.com/Sirupsen/logrus"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/lobby/format"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// major ver -> migration routine
var migrationRoutines = map[uint64]func(){
	2:  lobbyTypeChange,
	3:  dropSubtituteTable,
	4:  increaseChatMessageLength,
	5:  updateAllPlayerInfo,
	6:  truncateHTTPSessions,
	7:  setMumbleInfo,
	8:  setPlayerExternalLinks,
	9:  setPlayerSettings,
	10: dropTableSessions,
	11: dropColumnUpdatedAt,
}

func whitelist_id_string() {
	var count int

	db.DB.Model(&lobby.Lobby{}).Count(&count)
	if count == 0 {
		db.DB.Exec("ALTER TABLE lobbies DROP COLUMN whitelist")
		db.DB.Exec("ALTER TABLE lobbies ADD whitelist varchar(255)")
	}

	var whitelistIDs []int
	var lobbyIDs []uint

	db.DB.Model(&lobby.Lobby{}).Order("whitelist").Pluck("whitelist", &whitelistIDs)
	if len(whitelistIDs) == 0 {
		return
	}

	db.DB.Model(&lobby.Lobby{}).Order("id").Pluck("id", &lobbyIDs)

	db.DB.Exec("ALTER TABLE lobbies DROP whitelist")
	db.DB.Exec("ALTER TABLE lobbies ADD whitelist varchar(255)")

	for i, lobbyID := range lobbyIDs {
		db.DB.Model(&lobby.Lobby{}).Where("id = ?", lobbyID).Update("whitelist", strconv.Itoa(whitelistIDs[i]))
	}
}

func lobbyTypeChange() {
	newLobbyType := map[int]format.Format{
		6: format.Sixes,
		9: format.Highlander,
		4: format.Fours,
		3: format.Ultiduo,
		2: format.Bball,
		1: format.Debug,
	}

	var lobbyIDs []uint
	db.DB.Model(&lobby.Lobby{}).Order("id").Pluck("id", &lobbyIDs)

	for _, lobbyID := range lobbyIDs {
		var old int
		db.DB.DB().QueryRow("SELECT type FROM lobbies WHERE id = $1", lobbyID).Scan(&old)
		db.DB.Model(&lobby.Lobby{}).Where("id = ?", lobbyID).Update("type", newLobbyType[old])
	}
}

func dropSubtituteTable() {
	db.DB.Exec("DROP TABLE substitutes")
}

func increaseChatMessageLength() {
	db.DB.Exec("ALTER TABLE chat_messages ALTER COLUMN message TYPE character varying(150)")
}

func updateAllPlayerInfo() {
	var players []*player.Player
	db.DB.Model(&player.Player{}).Find(&players)

	for _, player := range players {
		player.UpdatePlayerInfo()
		player.Save()
	}
}

func truncateHTTPSessions() {
	db.DB.Exec("TRUNCATE TABLE http_sessions")
}

func setMumbleInfo() {
	var players []*player.Player

	db.DB.Model(&player.Player{}).Find(&players)
	for _, player := range players {
		player.MumbleUsername = strconv.Itoa(rand.Int())
		player.MumbleAuthkey = player.GenAuthKey()
		player.Save()
	}
}

func setPlayerExternalLinks() {
	var players []*player.Player
	db.DB.Model(&player.Player{}).Find(&players)

	for _, player := range players {
		player.ExternalLinks = make(postgres.Hstore)
		player.SetExternalLinks()
		player.Save()
	}
}

// move player_settings values to player.Settings hstore
func setPlayerSettings() {
	rows, err := db.DB.DB().Query("SELECT player_id, key, value FROM player_settings")
	if err != nil {
		logrus.Fatal(err)
	}
	for rows.Next() {
		var playerID uint
		var key, value string

		rows.Scan(&playerID, &key, &value)
		p, _ := player.GetPlayerByID(playerID)
		p.SetSetting(key, value)
	}

	db.DB.Exec("DROP TABLE player_settings")
}

func dropTableSessions() {
	db.DB.Exec("DROP TABLE http_sessions")
}

func dropColumnUpdatedAt() {
	db.DB.Exec("ALTER TABLE players DROP COLUMN updated_at")
}

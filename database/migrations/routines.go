// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"strconv"

	"github.com/Sirupsen/logrus"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
	"github.com/jinzhu/gorm"
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

	db.DB.Table("lobbies").Count(&count)
	if count == 0 {
		db.DB.Exec("ALTER TABLE lobbies DROP COLUMN whitelist")
		db.DB.Exec("ALTER TABLE lobbies ADD whitelist varchar(255)")
	}

	var whitelistIDs []int
	var lobbyIDs []uint

	db.DB.Table("lobbies").Order("whitelist").Pluck("whitelist", &whitelistIDs)
	if len(whitelistIDs) == 0 {
		return
	}

	db.DB.Table("lobbies").Order("id").Pluck("id", &lobbyIDs)

	db.DB.Exec("ALTER TABLE lobbies DROP whitelist")
	db.DB.Exec("ALTER TABLE lobbies ADD whitelist varchar(255)")

	for i, lobbyID := range lobbyIDs {
		db.DB.Model(&models.Lobby{}).Where("id = ?", lobbyID).Update("whitelist", strconv.Itoa(whitelistIDs[i]))
	}
}

func lobbyTypeChange() {
	newLobbyType := map[int]models.LobbyType{
		6: models.LobbyTypeSixes,
		9: models.LobbyTypeHighlander,
		4: models.LobbyTypeFours,
		3: models.LobbyTypeUltiduo,
		2: models.LobbyTypeBball,
		1: models.LobbyTypeDebug,
	}

	var lobbyIDs []uint
	db.DB.Table("lobbies").Order("id").Pluck("id", &lobbyIDs)

	for _, lobbyID := range lobbyIDs {
		var old int
		db.DB.DB().QueryRow("SELECT type FROM lobbies WHERE id = $1", lobbyID).Scan(&old)
		db.DB.Table("lobbies").Where("id = ?", lobbyID).Update("type", newLobbyType[old])
	}
}

func dropSubtituteTable() {
	db.DB.Exec("DROP TABLE substitutes")
}

func increaseChatMessageLength() {
	db.DB.Exec("ALTER TABLE chat_messages ALTER COLUMN message TYPE character varying(150)")
}

func updateAllPlayerInfo() {
	var players []*models.Player
	db.DB.Table("players").Find(&players)

	for _, player := range players {
		player.UpdatePlayerInfo()
		player.Save()
	}
}

func truncateHTTPSessions() {
	db.DB.Exec("TRUNCATE TABLE http_sessions")
}

func setMumbleInfo() {
	var players []*models.Player

	db.DB.Table("players").Find(&players)
	for _, player := range players {
		player.MumbleUsername = player.GenMumbleUsername()
		player.MumbleAuthkey = player.GenAuthKey()
		player.Save()
	}
}

func setPlayerExternalLinks() {
	var players []*models.Player
	db.DB.Table("players").Find(&players)

	for _, player := range players {
		player.ExternalLinks = make(gorm.Hstore)
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
		player, _ := models.GetPlayerByID(playerID)
		player.SetSetting(key, value)
	}

	db.DB.Exec("DROP TABLE player_settings")
}

func dropTableSessions() {
	db.DB.Exec("DROP TABLE http_sessions")
}

func dropColumnUpdatedAt() {
	db.DB.Exec("ALTER TABLE players DROP COLUMN updated_at")
}

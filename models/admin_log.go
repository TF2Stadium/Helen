package models

import (
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/jinzhu/gorm"
)

type AdminLogEntry struct {
	gorm.Model
	PlayerID uint
	Player   Player
	RelID    uint   `sql:"default:0"`
	RelText  string `sql:"default:''"`
}

func LogCustomAdminAction(playerid uint, reltext string, relid uint) error {
	entry := AdminLogEntry{
		PlayerID: playerid,
		RelID:    relid,
		RelText:  reltext,
	}

	return database.DB.Create(&entry).Error
}

func LogAdminAction(playerid uint, permission authority.AuthAction, relid uint) error {
	return LogCustomAdminAction(playerid, helpers.ActionNames[permission], relid)
}

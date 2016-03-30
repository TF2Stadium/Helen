// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	"github.com/Sirupsen/logrus"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/blang/semver"
)

//follows semantic versioning scheme
var schemaVersion = semver.Version{
	Major: 11,
	Minor: 0,
	Patch: 0,
}

type Constant struct {
	SchemaVersion string
}

func getCurrConstants() *Constant {
	constant := &Constant{}
	db.DB.Model(&Constant{}).Last(constant)

	return constant
}

func writeConstants() {
	db.DB.Exec("UPDATE constants SET schema_version = ?", schemaVersion.String())
	logrus.Info("Current Schema Version: ", getCurrConstants().SchemaVersion)
}

func checkSchema() {
	var count int
	defer writeConstants()

	db.DB.Model(&Constant{}).Where("schema_version = ?", schemaVersion.String()).Count(&count)

	if count == 1 {
		return
	}

	currStr := getCurrConstants().SchemaVersion
	if currStr == "" {
		db.DB.Save(&Constant{
			schemaVersion.String(),
		})
		//Initial database migration
		whitelist_id_string()
		//Write current schema version
		return
	}

	if v, _ := semver.Parse(currStr); v.Major < schemaVersion.Major {
		logrus.Warning("Incompatible schema change detected (", currStr, ") attempting to migrate to (", schemaVersion.String(), ")")
		for i := v.Major + 1; i <= schemaVersion.Major; i++ {
			logrus.Debug("Calling migration routine for ", schemaVersion.String())
			f := migrationRoutines[i]
			f()
		}
	}
}

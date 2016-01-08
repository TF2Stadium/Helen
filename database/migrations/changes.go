// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package migrations

import (
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/blang/semver"
)

//follows semantic versioning scheme
var schemaVersion = semver.Version{
	Major: 2,
	Minor: 0,
	Patch: 0,
}

type Constant struct {
	SchemaVersion string
}

func getCurrConstants() *Constant {
	constant := &Constant{}
	db.DB.Table("constants").Last(constant)

	return constant
}

func writeConstants() {
	db.DB.Exec("UPDATE constants SET schema_version = ?", schemaVersion.String())
	helpers.Logger.Info("Current Schema Version: %s", getCurrConstants().SchemaVersion)
}

func checkSchema() {
	var count int
	defer writeConstants()

	db.DB.Table("constants").Where("schema_version = ?", schemaVersion.String()).Count(&count)

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
		helpers.Logger.Warning("Incompatible schema change detected (%s), attempting to migrate to (%s).", currStr, schemaVersion.String())
		f := migrationRoutines[schemaVersion.Major]
		f()
	}
}

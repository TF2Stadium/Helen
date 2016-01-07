// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"testing"

	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	. "github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
	testhelpers.CleanupDB()
}

func TestLogCreation(t *testing.T) {
	t.Parallel()
	var obj = AdminLogEntry{}
	count := 5
	database.DB.Model(obj).Count(&count)
	assert.Equal(t, 0, count)

	LogAdminAction(1, helpers.ActionBanJoin, 2)
	LogCustomAdminAction(2, "test", 4)

	database.DB.Model(obj).Count(&count)
	assert.Equal(t, 2, count)
}

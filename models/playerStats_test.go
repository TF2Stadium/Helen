// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"testing"

	"github.com/TF2Stadium/Helen/database"
	_ "github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	. "github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	testhelpers.CleanupDB()
}

func TestLobbiesPlayed(t *testing.T) {
	t.Parallel()
	stats1 := &PlayerStats{}

	stats1.PlayedCountIncrease(LobbyTypeSixes) // sixes: 0 -> 1

	database.DB.Save(stats1)

	// can load the record
	var stats2 PlayerStats
	err := database.DB.First(&stats2, stats1.ID).Error
	assert.Nil(t, err)

	assert.Equal(t, 1, stats2.PlayedSixesCount)
}

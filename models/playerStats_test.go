// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"testing"

	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/testhelpers"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
}

func TestLobbiesPlayed(t *testing.T) {
	testhelpers.CleanupDB()
	stats1 := &models.PlayerStats{}

	stats1.PlayedCountSet(models.LobbyTypeSixes, 5)
	stats1.PlayedCountSet(models.LobbyTypeHighlander, 8)
	stats1.PlayedCountIncrease(models.LobbyTypeSixes) // sixes: 5 -> 6

	assert.Equal(t, 6, stats1.PlayedCountGet(models.LobbyTypeSixes))
	assert.Equal(t, 8, stats1.PlayedCountGet(models.LobbyTypeHighlander))
	database.DB.Save(stats1)

	// can load the record
	var stats2 models.PlayerStats
	err := database.DB.First(&stats2, stats1.ID).Error
	assert.Nil(t, err)

	assert.Equal(t, 6, stats2.PlayedCountGet(models.LobbyTypeSixes))
	assert.Equal(t, 8, stats2.PlayedCountGet(models.LobbyTypeHighlander))
}

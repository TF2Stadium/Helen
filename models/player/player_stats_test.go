// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package player_test

import (
	"testing"

	"github.com/TF2Stadium/Helen/database"
	_ "github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/lobby/format"
	. "github.com/TF2Stadium/Helen/models/player"
	"github.com/stretchr/testify/assert"
)

func TestLobbiesPlayed(t *testing.T) {
	t.Parallel()
	stats1 := &PlayerStats{}

	stats1.PlayedCountIncrease(format.Sixes) // sixes: 0 -> 1

	database.DB.Save(stats1)

	// can load the record
	var stats2 PlayerStats
	err := database.DB.First(&stats2, stats1.ID).Error
	assert.Nil(t, err)

	assert.Equal(t, 1, stats2.PlayedSixesCount)
}

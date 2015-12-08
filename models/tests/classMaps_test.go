// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"testing"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
}

func TestUgcHighlander(t *testing.T) {
	res, err := models.LobbyGetPlayerSlot(models.LobbyTypeSixes, "red", "scout1")
	assert.Equal(t, 0, res)
	assert.Nil(t, err)

	res, err = models.LobbyGetPlayerSlot(models.LobbyTypeHighlander, "blu", "heavy")
	assert.Equal(t, 13, res)
	assert.Nil(t, err)

	res, err = models.LobbyGetPlayerSlot(models.LobbyTypeHighlander, "blu", "garbageman")
	assert.NotNil(t, err)

	res, err = models.LobbyGetPlayerSlot(models.LobbyTypeSixes, "ylw", "demoman")
	assert.NotNil(t, err)
}

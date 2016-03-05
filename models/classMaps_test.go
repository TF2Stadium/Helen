// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models_test

import (
	"testing"

	_ "github.com/TF2Stadium/Helen/helpers"
	. "github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func TestClassMaps(t *testing.T) {
	res, err := LobbyGetPlayerSlot(LobbyTypeSixes, "red", "scout1")
	assert.Equal(t, 0, res)
	assert.Nil(t, err)

	res, err = LobbyGetPlayerSlot(LobbyTypeHighlander, "blu", "heavy")
	assert.Equal(t, 13, res)
	assert.Nil(t, err)

	res, err = LobbyGetPlayerSlot(LobbyTypeHighlander, "blu", "garbageman")
	assert.NotNil(t, err)

	res, err = LobbyGetPlayerSlot(LobbyTypeSixes, "ylw", "demoman")
	assert.NotNil(t, err)

	slots := []struct {
		n     int
		class string
	}{
		{0, "scout1"},
		{1, "scout2"},
		{2, "roamer"},
		{3, "pocket"},
		{4, "demoman"},
		{5, "medic"}}
	for _, slot := range slots {
		team, class, err := LobbyGetSlotInfoString(LobbyTypeSixes, slot.n)
		assert.NoError(t, err)
		assert.Equal(t, slot.class, class)
		assert.Equal(t, team, "red")
	}
}

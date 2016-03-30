// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package format_test

import (
	"testing"

	_ "github.com/TF2Stadium/Helen/helpers"
	. "github.com/TF2Stadium/Helen/models/lobby/format"
	"github.com/stretchr/testify/assert"
)

var slots = []struct {
	n     int
	class string
}{
	{0, "scout1"},
	{1, "scout2"},
	{2, "roamer"},
	{3, "pocket"},
	{4, "demoman"},
	{5, "medic"}}

func TestClassMaps(t *testing.T) {
	res, err := GetSlot(Sixes, "red", "scout1")
	assert.Equal(t, 0, res)
	assert.Nil(t, err)

	res, err = GetSlot(Highlander, "blu", "heavy")
	assert.Equal(t, 13, res)
	assert.Nil(t, err)

	res, err = GetSlot(Highlander, "blu", "garbageman")
	assert.NotNil(t, err)

	res, err = GetSlot(Sixes, "ylw", "demoman")
	assert.NotNil(t, err)

	for _, slot := range slots {
		team, class, err := GetSlotTeamClass(Sixes, slot.n)
		assert.NoError(t, err)
		assert.Equal(t, slot.class, class)
		assert.Equal(t, team, "red")
	}
}

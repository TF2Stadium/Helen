package controllerhelpers

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
	res, err := GetPlayerSlot(models.LobbyTypeSixes, "red", "scout1")
	assert.Equal(t, 0, res)
	assert.Nil(t, err)

	res, err = GetPlayerSlot(models.LobbyTypeHighlander, "blu", "heavy")
	assert.Equal(t, 13, res)
	assert.Nil(t, err)

	res, err = GetPlayerSlot(models.LobbyTypeHighlander, "blu", "garbageman")
	assert.NotNil(t, err)

	res, err = GetPlayerSlot(models.LobbyTypeSixes, "ylw", "demoman")
	assert.NotNil(t, err)
}

package models_test

import (
	"strconv"
	"testing"
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
}

func TestNewChatMessage(t *testing.T) {
	testhelpers.CleanupDB()

	lobby := testhelpers.CreateLobby()
	lobby.Save()

	player := testhelpers.CreatePlayer()
	player.Save()

	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)

		message, _ := models.NewChatMessage(strconv.Itoa(i), 0, player)
		assert.NotNil(t, message)

		err := db.DB.Save(message).Error
		assert.Nil(t, err)
	}

	messages, err := models.GetRoomMessages(0)
	assert.Nil(t, err)
	assert.Equal(t, len(messages), 3)

	for i := 1; i < len(messages); i++ {
		assert.True(t, messages[i].CreatedAt.Unix() > messages[i-1].CreatedAt.Unix())
	}

	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)

		message, _ := models.NewChatMessage(strconv.Itoa(i), 1, player)
		assert.NotNil(t, message)

		err := db.DB.Save(message).Error
		assert.Nil(t, err)
	}

	messages, err = models.GetPlayerMessages(player)
	assert.Nil(t, err)
	assert.Equal(t, len(messages), 6)

	for i := 1; i < len(messages); i++ {
		assert.True(t, messages[i].CreatedAt.Unix() > messages[i-1].CreatedAt.Unix())
	}

	messages, err = models.GetPlayerMessages(player)
	assert.Nil(t, err)
	assert.Equal(t, len(messages), 6)
}

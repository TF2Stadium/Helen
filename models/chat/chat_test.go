package chat_test

import (
	"strconv"
	"testing"

	db "github.com/TF2Stadium/Helen/database"
	_ "github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/internal/testhelpers"
	. "github.com/TF2Stadium/Helen/models/chat"
	"github.com/stretchr/testify/assert"
)

func init() {
	testhelpers.CleanupDB()
}

func TestNewChatMessage(t *testing.T) {
	lobby := testhelpers.CreateLobby()
	defer lobby.Close(false, true)
	lobby.Save()

	player := testhelpers.CreatePlayer()
	player.Save()

	for i := 0; i < 3; i++ {
		message := NewChatMessage(strconv.Itoa(i), 0, player)
		assert.NotNil(t, message)

		err := db.DB.Save(message).Error
		assert.Nil(t, err)
	}

	messages, err := GetRoomMessages(0)
	assert.Nil(t, err)
	assert.Equal(t, len(messages), 3)
}

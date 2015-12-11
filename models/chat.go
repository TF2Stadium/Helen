package models

import (
	"time"

	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
)

type ChatMessage struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"-"`

	Timestamp int64 `sql:"-" json:"timestamp"`

	PlayerID uint          `json:"-"`
	Player   PlayerSummary `json:"player" sql:"-"`

	Room    int    `json:"room"`
	Message string `json:"message" sql:"type:varchar(120)"`
	Deleted bool   `json:"-"`
}

func NewChatMessage(message string, room int, player *Player) (*ChatMessage, *helpers.TPError) {
	if banned, _ := player.IsBannedWithTime(PlayerBanChat); banned {
		return nil, helpers.NewTPError("Player has been chat-banned.", 2)
	}

	record := &ChatMessage{
		Timestamp: time.Now().Unix(),

		PlayerID: player.ID,
		Player:   DecoratePlayerSummary(player),

		Room:    room,
		Message: message,
	}

	return record, nil
}

func GetRoomMessages(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("room = ?", room).Order("created_at").Find(&messages).Error

	return messages, err
}

//Get all messages sent by player in a specfified room
func GetPlayerMessages(player *Player) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("player_id = ?", player.ID).Order("room, created_at").Find(&messages).Error

	return messages, err

}

func GetScrollback(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("room = ?", room).Order("id desc").Limit(20).Find(&messages).Error

	for _, message := range messages {
		var player Player
		db.DB.First(&player, message.PlayerID)
		message.Player = DecoratePlayerSummary(&player)
		message.Timestamp = message.CreatedAt.Unix()
	}
	return messages, err
}

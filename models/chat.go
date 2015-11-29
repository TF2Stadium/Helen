package models

import (
	"time"

	db "github.com/TF2Stadium/Helen/database"
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

func NewChatMessage(message string, room int, player *Player) *ChatMessage {
	return &ChatMessage{
		Timestamp: time.Now().Unix(),

		PlayerID: player.ID,
		Player:   DecoratePlayerSummary(player),

		Room:    room,
		Message: message,
	}
}

func GetMessages(player *Player, room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("player_id = ? AND room = ?", player.ID, room).Order("created_at").Find(&messages).Error

	return messages, err
}

//Get all messages sent by player in a specfified room
func GetAllMessages(player *Player) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("player_id = ?", player.ID).Order("created_at").Find(&messages).Error

	return messages, err

}

func GetScrollback(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("room = ?", room).Order("id desc").Limit(20).Find(&messages).Error

	for _, message := range messages {
		var player Player
		db.DB.First(&player, message.PlayerID)
		message.Player = DecoratePlayerSummary(&player)
	}
	return messages, err
}

package models

import (
	"fmt"
	"time"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
)

// ChatMessage Represents a chat mesasge sent by a particular player
type ChatMessage struct {
	// Message ID
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"-"`

	// Because the frontend needs the unix timestamp for the message. Not stored in the DB
	Timestamp int64 `sql:"-" json:"timestamp"`

	// ID of the player who sent the message
	PlayerID uint `json:"-"`
	// Not in the DB, used by frontend to retrieve player information
	Player PlayerSummary `json:"player" sql:"-"`

	// Room to which the message was sent
	Room int `json:"room"`
	// The actual Message, limited to 120 characters
	Message string `json:"message" sql:"type:varchar(120)"`
	// True if the message has been deleted by a moderator
	Deleted bool `json:"-"`
	// true if the message is sent by a bot
	Bot bool `json:"bot"`
}

// Return a new ChatMessage sent from specficied player
func NewChatMessage(message string, room int, player *Player) *ChatMessage {
	record := &ChatMessage{
		Timestamp: time.Now().Unix(),

		PlayerID: player.ID,
		Player:   DecoratePlayerSummary(player),

		Room:    room,
		Message: message,
	}

	return record
}

func (m *ChatMessage) Save() { db.DB.Save(m) }

func (m *ChatMessage) Send(room int) {
	broadcaster.SendMessageToRoom(fmt.Sprintf("%d_public", room), "chatReceive", m)
	if room != 0 {
		broadcaster.SendMessageToRoom(fmt.Sprintf("%d_private", room), "chatReceive", m)
	}
}

func NewBotMessage(message string, room int) *ChatMessage {
	m := &ChatMessage{
		Timestamp: time.Now().Unix(),

		Player:  PlayerSummary{Name: "TF2Stadium"},
		Room:    room,
		Message: message,

		Bot: true,
	}

	m.Save()
	return m
}

func SendNotification(message string, room int) {
	pub := fmt.Sprintf("%d_public", room)
	broadcaster.SendMessageToRoom(pub, "chatReceive", NewBotMessage(message, room))
}

// Return a list of ChatMessages spoken in room
func GetRoomMessages(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("room = ?", room).Order("created_at").Find(&messages).Error

	return messages, err
}

// Return all messages sent by player to room
func GetPlayerMessages(player *Player) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("player_id = ?", player.ID).Order("room, created_at").Find(&messages).Error

	return messages, err

}

// Get a list of last 20 messages sent to room, used by frontend for displaying the chat history/scrollback
func GetScrollback(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("room = ?", room).Order("id desc").Limit(20).Find(&messages).Error

	for _, message := range messages {
		var player Player
		if message.Bot {
			message.Player = PlayerSummary{Name: "TF2Stadium"}
		} else {
			db.DB.First(&player, message.PlayerID)
			message.Player = DecoratePlayerSummary(&player)
		}
		message.Timestamp = message.CreatedAt.Unix()
	}
	return messages, err
}

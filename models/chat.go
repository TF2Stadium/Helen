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
	CreatedAt time.Time `json:"timestamp"`

	Player   Player `json:"player"` // not in the DB, used by frontend to retrieve player information
	PlayerID uint   `json:"-"`      // ID of the player who sent the message

	// Room to which the message was sent
	Room int `json:"room"`
	// The actual Message, limited to 150 characters
	Message string `json:"message" sql:"type:varchar(150)"`
	// True if the message has been deleted by a moderator
	Deleted bool `json:"deleted"`
	// true if the message is sent by a bot
	Bot bool `json:"bot"`
	// true if the message is in-game
	InGame bool `json:"ingame"`
}

var botSummary = Player{
	JSONFields: JSONFields{
		PlaceholderTags: new([]string),
	},
}

func init() {
	botSummary.Name = "TF2Stadium"
	*botSummary.PlaceholderTags = []string{"tf2stadium"}
}

// Return a new ChatMessage sent from specficied player
func NewChatMessage(message string, room int, player *Player) *ChatMessage {
	player.SetPlayerSummary()
	record := &ChatMessage{
		PlayerID: player.ID,
		Player:   *player,

		Room:    room,
		Message: message,
	}

	return record
}

func NewInGameChatMessage(lobby *Lobby, player *Player, message string) *ChatMessage {
	return &ChatMessage{
		PlayerID: player.ID,
		Player:   *player,

		Room:    int(lobby.ID),
		Message: message,
		InGame:  true,
	}
}

func (m *ChatMessage) Save() { db.DB.Save(m) }

func (m *ChatMessage) Send() {
	broadcaster.SendMessageToRoom(fmt.Sprintf("%d_public", m.Room), "chatReceive", m)
	if m.Room != 0 {
		broadcaster.SendMessageToRoom(fmt.Sprintf("%d_private", m.Room), "chatReceive", m)
	}
}

func NewBotMessage(message string, room int) *ChatMessage {
	m := &ChatMessage{
		Player:  botSummary,
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

	err := db.DB.Table("chat_messages").Where("room = ? AND deleted = FALSE", room).Order("id desc").Limit(20).Find(&messages).Error

	for _, message := range messages {
		var player Player
		if message.Bot {
			message.Player = botSummary
		} else {
			db.DB.First(&player, message.PlayerID)
			message.Player = player
		}
	}
	return messages, err
}

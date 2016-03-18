package models

import (
	"fmt"
	"time"

	"encoding/json"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
)

// ChatMessage Represents a chat mesasge sent by a particular player
type ChatMessage struct {
	// Message ID
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"timestamp"`
	Player    Player    `json:"-"`
	PlayerID  uint      `json:"-"`                               // ID of the player who sent the message
	Room      int       `json:"room"`                            // room to which the message was sent
	Message   string    `json:"message" sql:"type:varchar(150)"` // the actual Message
	Deleted   bool      `json:"deleted"`                         // true if the message has been deleted by a moderator
	Bot       bool      `json:"bot"`                             // true if the message was sent by the notification "bot"
	InGame    bool      `json:"ingame"`                          // true if the message is in-game
}

// Return a new ChatMessage sent from specficied player
func NewChatMessage(message string, room int, player *Player) *ChatMessage {
	player.SetPlayerSummary()
	record := &ChatMessage{
		PlayerID: player.ID,

		Room:    room,
		Message: message,
	}

	return record
}

func NewInGameChatMessage(lobby *Lobby, player *Player, message string) *ChatMessage {
	return &ChatMessage{
		PlayerID: player.ID,

		Room:    int(lobby.ID),
		Message: message,
		InGame:  true,
	}
}

func (m *ChatMessage) Save() { db.DB.Save(m) }

func (m *ChatMessage) Send() {
	broadcaster.SendMessageToRoom(fmt.Sprintf("%d_public", m.Room), "chatReceive", (*sentMessage)(m))
	if m.Room != 0 {
		broadcaster.SendMessageToRoom(fmt.Sprintf("%d_private", m.Room), "chatReceive", (*sentMessage)(m))
	}
}

// we only need these three things for showing player messages
type minPlayer struct {
	Name    string   `json:"name"`
	SteamID string   `json:"steamid"`
	Tags    []string `json:"tags"`
}

// sentMessage aliases ChatMessage, since implementing MarshalJSON for
// ChatMessage would result in a recursive data structure (see below),
// which the json parser cannot marshal
type sentMessage ChatMessage

func (m *sentMessage) MarshalJSON() ([]byte, error) {
	var player *Player

	if !m.Bot {
		player, _ = GetPlayerByID(m.PlayerID)
	}

	message := struct {
		*ChatMessage
		Player *minPlayer `json:"player"`
	}{(*ChatMessage)(m), &minPlayer{}}

	if m.Bot {
		message.Player.Name = "TF2Stadium"
		message.Player.Tags = []string{"tf2stadium"}
	} else {
		message.Player.Name = player.Alias()
		message.Player.SteamID = player.SteamID
		message.Player.Tags = decoratePlayerTags(player)
	}

	if m.Deleted {
		message.Message = "<deleted>"
		message.Player.Tags = append(message.Player.Tags, "deleted")
	}

	return json.Marshal(message)
}

func NewBotMessage(message string, room int) *ChatMessage {
	m := &ChatMessage{
		Room:    room,
		Message: message,

		Bot: true,
	}

	m.Save()
	return m
}

func SendNotification(message string, room int) {
	pub := fmt.Sprintf("%d_public", room)
	broadcaster.SendMessageToRoom(pub, "chatReceive", (*sentMessage)(NewBotMessage(message, room)))
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
func GetScrollback(room int) ([]*sentMessage, error) {
	var messages []*sentMessage // apparently the ORM works fine with using this type (they're aliases after all)

	err := db.DB.Table("chat_messages").Where("room = ? AND deleted = FALSE", room).Order("id desc").Limit(20).Find(&messages).Error

	return messages, err
}

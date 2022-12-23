package chat

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models/player"
)

// ChatMessage Represents a chat mesasge sent by a particular player
type ChatMessage struct {
	// Message ID
	ID        uint          `json:"id"`
	CreatedAt time.Time     `json:"timestamp"`
	Player    player.Player `json:"-"`
	PlayerID  uint          `json:"-"`                               // ID of the player who sent the message
	Room      int           `json:"room"`                            // room to which the message was sent
	Message   string        `json:"message" sql:"type:varchar(150)"` // the actual Message
	Deleted   bool          `json:"deleted"`                         // true if the message has been deleted by a moderator
	Bot       bool          `json:"bot"`                             // true if the message was sent by the notification "bot"
	InGame    bool          `json:"ingame"`                          // true if the message is in-game
}

// Return a new ChatMessage sent from specficied player
func NewChatMessage(message string, room int, player *player.Player) *ChatMessage {
	player.SetPlayerSummary()
	record := &ChatMessage{
		PlayerID: player.ID,

		Room:    room,
		Deleted: filteredMessage(message),
		Message: message,
	}

	return record
}

func NewInGameChatMessage(lobbyID uint, player *player.Player, message string) *ChatMessage {
	return &ChatMessage{
		PlayerID: player.ID,

		Room:    int(lobbyID),
		Message: message,
		Deleted: filteredMessage(message),
		InGame:  true,
	}
}

func (m *ChatMessage) Save() {
	if !m.Bot {
		var count int
		db.DB.Table("chat_messages").
			Where("player_id = ? AND timestamp >= ?",
				m.PlayerID,
				time.Now().Add(-1*config.Constants.ChatRateLimit)).
			Count(&count)
		if count > 0 {
			return
		}

	}
	db.DB.Save(m)
}

func (m *ChatMessage) Send() {
	broadcaster.SendMessageToRoom(fmt.Sprintf("%d_public", m.Room), "chatReceive", m)
	if m.Room != 0 {
		broadcaster.SendMessageToRoom(fmt.Sprintf("%d_private", m.Room), "chatReceive", m)
	}
}

// we only need these three things for showing player messages
type minPlayer struct {
	Name    string   `json:"name"`
	SteamID string   `json:"steamid"`
	Tags    []string `json:"tags"`
}

var bot = minPlayer{"TF2Stadium", "76561198275497635", []string{"tf2stadium"}}

func (m *ChatMessage) MarshalJSON() ([]byte, error) {
	message := map[string]interface{}{
		"id":        m.ID,
		"timestamp": m.CreatedAt,
		"room":      m.Room,
		"message":   m.Message,
		"deleted":   m.Deleted,
		"ingame":    m.InGame,
	}
	if m.Bot {
		message["player"] = bot
	} else {
		p := &player.Player{}
		db.DB.First(p, m.PlayerID)
		player := minPlayer{
			Name:    p.Alias(),
			SteamID: p.SteamID,
			Tags:    p.DecoratePlayerTags(),
		}

		if m.Deleted {
			player.Tags = append(player.Tags, "<deleted>")
			message["message"] = "<deleted>"
		}

		message["player"] = player
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
	broadcaster.SendMessageToRoom(pub, "chatReceive", NewBotMessage(message, room))
}

// Return a list of ChatMessages spoken in room
func GetRoomMessages(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Model(&ChatMessage{}).Where("room = ?", room).Order("created_at").Find(&messages).Error

	return messages, err
}

// Return all messages sent by player to room
func GetPlayerMessages(p *player.Player) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Model(&ChatMessage{}).Where("player_id = ?", p.ID).Order("room, created_at").Find(&messages).Error

	return messages, err

}

// Get a list of last 20 messages sent to room, used by frontend for displaying the chat history/scrollback
func GetScrollback(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage // apparently the ORM works fine with using this type (they're aliases after all)

	err := db.DB.Table("chat_messages").Where("room = ? AND deleted = FALSE", room).Order("id desc").Limit(20).Find(&messages).Error

	return messages, err
}

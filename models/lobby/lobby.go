package lobby

import (
	"errors"
	"log"
	"time"

	"github.com/TeamPlayTF/Server/database"
	"github.com/TeamPlayTF/Server/helpers"
	"github.com/TeamPlayTF/Server/models"
	"gopkg.in/mgo.v2/bson"
)

type LobbyType int
type Whitelist int
type LobbyState int

const (
	LobbyTypeSixes      LobbyType = 0
	LobbyTypeHighlander LobbyType = 1
)

const (
	LobbyStateWaiting    LobbyState = 0
	LobbyStateInProgress LobbyState = 1
	LobbyStateEnded      LobbyState = 2
)

var typePlayerCount = map[LobbyType]int{
	LobbyTypeSixes:      6,
	LobbyTypeHighlander: 9,
}

//Given Lobby IDs are unique, we'll use them for mumble channel names
type Lobby struct {
	Id      bson.ObjectId `bson:"_id,omitempty"` //Lobby id
	MapName string
	State   LobbyState
	Type    LobbyType

	// Dependencies are objectId's
	PlayerIds []string
	ServerId  bson.ObjectId `bson:",omitempty"` // server id
	Whitelist Whitelist     //whitelist.tf ID
	CreatedAt time.Time
}

//id should be maintained in the main loop
func New(mapName string, lobbyType LobbyType /*server *Server,*/, whitelist int) *Lobby {
	lobby := &Lobby{
		Type:      lobbyType,
		State:     LobbyStateWaiting,
		MapName:   mapName,
		PlayerIds: make([]string, 2*typePlayerCount[lobbyType]),
		//Server:    server,
		Whitelist: Whitelist(whitelist), // that's a strange line
		CreatedAt: bson.Now(),
	}

	return lobby
}

func (lobby *Lobby) Save() error {
	if !lobby.Id.Valid() {
		lobby.Id = bson.NewObjectId()
	}
	_, err := database.GetCollection("lobbies").UpsertId(lobby.Id, lobby)
	return err
}

func (lobby *Lobby) GetPlayerObjects() []*models.Player {
	var ids []bson.ObjectId

	for _, playerID := range lobby.PlayerIds {
		if playerID != "" {
			ids = append(ids, bson.ObjectId(playerID))
		}
	}

	var result []*models.Player
	database.GetPlayersCollection().Find(bson.M{"_id": ids}).All(&result)

	return result
}

func (lobby *Lobby) GetPlayerSlot(player *models.Player) (slot int, err error) {
	err = errors.New("Player not in lobby")
	if !player.Id.Valid() {
		return -1, err
	}
	for playerSlot, playerId := range lobby.PlayerIds {
		if bson.ObjectId(playerId) == player.Id {
			return playerSlot, nil
		}

	}
	return -1, err
}

// // //Add player to lobby
func (lobby *Lobby) AddPlayer(player *models.Player, slot int) *helpers.TPError {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */

	badSlotError := helpers.NewTPError("This slot does not exist.", 3)
	filledError := helpers.NewTPError("This slot has been filled.", 2)
	alreadyInLobbyError := helpers.NewTPError("Player is already in a lobby", 1)

	if !player.Id.Valid() {
		return helpers.NewTPError("Player not in the database", -1)
	}

	if slot >= len(lobby.PlayerIds) {
		return badSlotError
	}

	playerLobbyId, err := player.InLobby()
	log.Println(err)

	if err == nil {
		//player is already in a lobby
		if playerLobbyId == lobby.Id { //lobby hasn't changed

			if bson.ObjectId(lobby.PlayerIds[slot]).Valid() {
				return filledError
			}

			// assign the player to a new slot
			oldSlot, _ := lobby.GetPlayerSlot(player)
			lobby.PlayerIds[slot] = string(player.Id)
			lobby.PlayerIds[oldSlot] = ""

			return nil
		}
		return alreadyInLobbyError
	}

	if lobby.PlayerIds[slot] != "" {
		//slot has been filled
		return filledError
	}

	lobby.PlayerIds[slot] = string(player.Id)

	return nil
}

func (lobby *Lobby) allSlotsFilled() bool {
	for _, playerId := range lobby.PlayerIds {
		if !bson.ObjectId(playerId).Valid() { //Slots haven't been filled
			return false
		}
	}
	return true
}

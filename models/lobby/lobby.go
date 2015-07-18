package lobby

import (
	"errors"
	"log"
	"time"

	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/TF2Stadium/Server/models"
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
	Room string //socket.io room for lobby.
	PlayerIds []string
	ServerId  bson.ObjectId `bson:",omitempty"` // server id
	Whitelist Whitelist     //whitelist.tf ID
	CreatedAt time.Time

	BannedPlayers map[bson.ObjectId]bool
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

func (lobby *Lobby) GetPlayerObjects() ([]*models.Player, error) {
	var ids []bson.ObjectId

	for _, playerID := range lobby.PlayerIds {
		if playerID != "" {
			ids = append(ids, bson.ObjectId(playerID))
		}
	}

	var result []*models.Player
	err := database.GetPlayersCollection().Find(bson.M{"_id": ids}).All(&result)

	return result, err
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

func GetLobbyById(id string) (*Lobby, *helpers.TPError) {
	invalidHex := helpers.NewTPError("lobbyid is not a valid hex representation", -2)
	nonExistentLobby := helpers.NewTPError("Lobby not in the database", -1) 

	if !bson.IsObjectIdHex(id) {
		return nil, invalidHex
	}
	
	lob := &Lobby{}
	lobbyid := bson.ObjectIdHex(id)
	err := database.GetLobbiesCollection().Find(lobbyid).One(lob)

	if err != nil {
		return nil, nonExistentLobby
	}

	return lob, nil
}

// // //Add player to lobby
func (lobby *Lobby) AddPlayer(player *models.Player, slot int) *helpers.TPError {
	/* Possible errors while joining
	 * Slot has been filled
	 * Player has already joined a lobby
	 * anything else?
	 */

	lobbyBanError := helpers.NewTPError("The player has been banned from this lobby.", 4)
	badSlotError := helpers.NewTPError("This slot does not exist.", 3)
	filledError := helpers.NewTPError("This slot has been filled.", 2)
	alreadyInLobbyError := helpers.NewTPError("Player is already in a lobby", 1)

	if !player.Id.Valid() {
		return helpers.NewTPError("Player not in the database", -1)
	}

	if _, ok := lobby.BannedPlayers[player.Id]; ok {
		return lobbyBanError
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

func (lobby *Lobby) RemovePlayer(player *models.Player) *helpers.TPError {
	slot, err := lobby.GetPlayerSlot(player)

	if err != nil {
		return helpers.NewTPError("Player not in any lobby.", 4)
	}
	lobby.PlayerIds[slot] = ""
	return nil
}

func (lobby *Lobby) KickAndBanPlayer(player *models.Player) *helpers.TPError {
	lobby.BannedPlayers[player.Id] = true
	return lobby.RemovePlayer(player)
}

func (lobby *Lobby) ReadyPlayer(player *models.Player) *helpers.TPError {
	// TODO implement
	return nil
}

func (lobby *Lobby) UnreadyPlayer(player *models.Player) *helpers.TPError {
	// TODO implement
	return nil
}

func (lobby *Lobby) IsPlayerReady(player *models.Player) (bool, *helpers.TPError) {
	// TODO implement
	return false, nil
}

func (lobby *Lobby) IsStarted() (bool, *helpers.TPError) {
	// TODO implement
	return false, nil
}

func (lobby *Lobby) AddSpectator(player *models.Player) *helpers.TPError {
	// TODO implement
	return nil
}

func (lobby *Lobby) RemoveSpectator(player *models.Player) *helpers.TPError {
	// TODO implement
	return nil
}

func (lobby *Lobby) GetSpectatorObjects() ([]*models.Player, *helpers.TPError) {
	// TODO implement
	return nil, nil
}

func (lobby *Lobby) IsFull() bool {
	// TODO broken, fix
	for _, playerId := range lobby.PlayerIds {
		if !bson.ObjectId(playerId).Valid() { //Slots haven't been filled
			return false
		}
	}
	return true
}

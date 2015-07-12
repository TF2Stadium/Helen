package lobby

import (
	"errors"
	"time"

	"github.com/TeamPlayTF/Server/database"
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
	Teams     [][]string    // RED - team[0], BLU - team[1]
	ServerId  bson.ObjectId `bson:",omitempty"` // server id
	Whitelist Whitelist     //whitelist.tf ID
	CreatedAt time.Time
}

//id should be maintained in the main loop
func New(mapName string, lobbyType LobbyType /*server *Server,*/, whitelist int) *Lobby {
	lobby := &Lobby{
		Type:    lobbyType,
		State:   LobbyStateWaiting,
		MapName: mapName,
		Teams:   make([][]string, 2),
		//Server:    server,
		Whitelist: Whitelist(whitelist), // that's a strange line
		CreatedAt: bson.Now(),
	}

	count := typePlayerCount[lobbyType]

	lobby.Teams[0] = make([]string, count)
	lobby.Teams[1] = make([]string, count)

	return lobby
}

func (lobby *Lobby) Save() error {
	if !lobby.Id.Valid() {
		lobby.Id = bson.NewObjectId()
	}
	_, err := database.GetCollection("lobbies").UpsertId(lobby.Id, lobby)
	return err
}

func (lobby *Lobby) Update() {
	if !lobby.Id.Valid() {
		return
	}

	database.GetCollection("lobbies").FindId(lobby.Id).One(lobby)
}

func (lobby *Lobby) GetPlayerObjects() []*models.Player {
	var ids []bson.ObjectId

	for _, team := range lobby.Teams {
		for _, playerID := range team {
			if playerID != "" {
				ids = append(ids, bson.ObjectId(playerID))
			}
		}
	}

	var result []*models.Player
	database.GetCollection("lobbies").Find(bson.M{"_id": ids}).All(&result)

	return result
}

func (lobby *Lobby) End() {
	lobby.State = LobbyStateEnded

	// notify players of this
}

func (lobby *Lobby) GetPlayerSlot(player *models.Player) (team, slot int, err error) {
	err = errors.New("Player not in lobby")
	if !player.Id.Valid() {
		return 0, 0, err
	}
	for teamIndex, team := range lobby.Teams {
		for playerSlot, playerId := range team {
			if bson.ObjectId(playerId) == player.Id {
				return teamIndex, playerSlot, nil
			}
		}
	}
	return 0, 0, err
}

// // //Add player to lobby
// func (lobby *Lobby) AddPlayer(player *Player, team int, slot int) error {
// 	/* Possible errors while joining
// 	 * Slot has been filled
// 	 * Player has already joined a lobby
// 	 * anything else?
// 	 */
//
// 	playerLobbyId := player.LobbyId
// 	filledError := helpers.NewTPError("This slot has been filled.", 2)
// 	alreadyInLobbyError := helpers.NewTPError("Player is already in a lobby", 1)
//
// 	if playerLobbyId.Valid() {
// 		//player is already in a lobby
// 		if playerLobbyId == lobby.Id { //lobby hasn't changed
//
// 			if lobby.Teams[team][slot].Valid() {
// 				return filledError
// 			}
//
// 			// assign the player to a new slot
// 			oldTeam, oldSlot, _ := lobby.GetPlayerSlot(player)
// 			lobby.Teams[team][slot] = player.Id
// 			lobby.Teams[oldTeam][oldSlot] = bson.ObjectId(0)
//
// 			return nil
// 		}
// 		return alreadyInLobbyError
// 	}
//
// 	if lobby.Teams[team][slot].Id.Valid() {
// 		//slot has been filled
// 		return filledError
// 	}
//
// 	lobby.Teams[team][slot] = player.Id
//
// 	//Check if all slots have been filled
//
// 	if lobby.allSlotsFilled() {
// 		/* Slots have been filled. Ask frontend to create the "Ready" button.
// 		 */
//
// 	}
// 	return nil
// }
//
// func (lobby *Lobby) allSlotsFilled() bool {
// 	for _, team := range lobby.Teams {
// 		for _, playerId := range team {
// 			if !playerId.Valid() { //Slots haven't been filled
// 				return false
// 			}
// 		}
// 	}
// 	return true
// }

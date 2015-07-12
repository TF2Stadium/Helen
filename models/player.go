package models

import (
	"errors"
	"time"

	"github.com/TeamPlayTF/Server/database"

	"gopkg.in/mgo.v2/bson"
)

type Player struct {
	Id        bson.ObjectId `bson:"_id,omitempty"` // MongoDB ID
	SteamId   string        // Players steam ID
	Name      string        // Player name
	CreatedAt time.Time     // Account creation time (not steam)
}

func NewPlayer(steamId string) *Player {
	player := &Player{SteamId: steamId}

	// magically get the player's name, avatar and other stuff from steam

	return player
}

func (player *Player) Save() error {
	if !player.Id.Valid() {
		player.Id = bson.NewObjectId()
	}
	_, err := database.GetPlayersCollection().UpsertId(player.Id, player)
	return err
}

func (player *Player) InLobby() (bson.ObjectId, error) {
	if !player.Id.Valid() {
		return "", errors.New("Player is not in the database")
	}

	type idStruct struct {
		Id bson.ObjectId
	}

	res := &idStruct{}
	err := database.GetLobbiesCollection().
		Find(bson.M{"playerids": string(player.Id)}).Select(bson.M{"_id": true}).One(res)

	if err != nil {
		return "", err
	}
	return res.Id, nil
}

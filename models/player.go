package models

import (
	"fmt"
	"github.com/TeamPlayTF/Server/database"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Player struct {
	Id      bson.ObjectId `bson:"_id,omitempty"        json:"id"`      // MongoDB ID
	SteamId string        `bson:"steamid,omitempty"    json:"steamid"` // Players steam ID
	Name    string        `bson:"name,omitempty"       json:"name"`    // Player name
	Created time.Time     `bson:"created,omitempty"    json:"created"` // Account creation time (not steam)
}

func NewPlayer() *Player {
	return new(Player)
}

func (p *Player) Create() error {
	fmt.Println("[Player]: Creating player from steamid -> [" + p.SteamId + "]")

	// should get player info from API, then
	apiErr := p.GetInfoFromAPI()
	if apiErr != nil {
		return apiErr
	}

	session, collection := database.Get("players")
	defer session.Close()

	// wrap user data into bson model
	userData := bson.M{
		"steamid": p.SteamId,
		"time":    time.Now(),
		"name":    p.Name,
		"_id":     bson.NewObjectId(),
	}

	// insert document into collection
	insErr := collection.Insert(userData)

	if insErr == nil {
		fmt.Println("[Player]: Player created!")
		p.Find()
	}

	return insErr
}

// TODO: should get player info from SteamAPI
func (p *Player) GetInfoFromAPI() error {
	fmt.Println("certain day, a lazy guy didn't do his job, instead he throws bananas at people")

	return nil
}

func (p *Player) Find() error {
	fmt.Println("[Player]: Getting player from steamid -> [" + p.SteamId + "]")

	session, collection := database.Get("players")
	defer session.Close()

	// will find a steamid with a limit of 1 doc
	// and insert into "p" which is the current *Player struct (sets itself into it)
	err := collection.
		Find(bson.M{"steamid": p.SteamId}).
		One(&p)

	return err
}

func (p *Player) Exists() (error, bool) {
	fmt.Println("[Player]: Checking if player exists")

	session, collection := database.Get("players")
	defer session.Close()

	exists := false
	var err error

	if p.SteamId != "" {
		err = collection.
			Find(bson.M{"steamid": p.SteamId}).
			One(&p)

	} else if p.Id.Hex() != "" {
		err = collection.
			Find(bson.M{"_id": p.Id}).
			One(&p)
	}

	if err == nil {
		exists = true
	} else {
		exists = false
	}

	return err, exists
}

func (p *Player) Delete() error {
	fmt.Println("[Player]: Removing player")

	session, collection := database.Get("players")
	defer session.Close()

	err, exists := p.Exists()

	if exists {
		if p.SteamId != "" {
			err = collection.Remove(bson.M{"steamid": p.SteamId})

		} else if p.Id.Hex() != "" {
			err = collection.Remove(bson.M{"_id": p.Id})
		}
	}

	return err
}

///////////////////
// Sets and Gets //
///////////////////

// sets

// TODO: add length checking
func (p *Player) SetSteamId(steamId string) {
	p.SteamId = steamId
}

func (p *Player) SetName(name string) {
	p.Name = name
}

// gets

func (p *Player) GetId() string {
	return p.Id.Hex()
}

func (p *Player) GetCreated() string {
	return p.Created.String()
}

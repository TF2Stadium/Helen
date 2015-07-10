package database

import (
	"fmt"
	"github.com/TeamPlayTF/Server/config"
	"github.com/TeamPlayTF/Server/database"
	"github.com/TeamPlayTF/Server/models"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
	"testing"
)

// should always be the 1st test (like a setup)
func TestDatabase(t *testing.T) {
	// start the database connection
	config.SetupConstants()
	database.Test()
	fmt.Println("[Test.Database] IsTest? " + strconv.FormatBool(database.IsTest))
	database.Init()

	// check if user exists then remove it
	// just to be sure we won't get any errors
	fErr := database.Database.C("players").Find(bson.M{"steamid": "76561198074578368"})
	if fErr == nil {
		rErr := database.Database.C("players").Remove(bson.M{"steamid": "76561198074578368"})
		if rErr != nil {
			log.Fatal(rErr)
		}
	}
}

// test the creation of a player
func TestDatabaseCreate(t *testing.T) {
	player := models.NewPlayer()

	fmt.Println("[Test.Player]: creating -> 76561198074578368")

	// steamid, name
	player.SetName("dunno his name")
	player.SetSteamId("76561198074578368") // some steamrep admin
	err := player.Create()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("[Test.Player.Create]: id -> " + player.GetId())
		fmt.Println("[Test.Player.Create]: name -> " + player.Name)
		fmt.Println("[Test.Player.Create]: steamid -> " + player.SteamId)
		fmt.Println("[Test.Player.Create]: created -> " + player.GetCreated())
	}
}

// test if the player exists
func TestDatabaseExists(t *testing.T) {
	player := models.NewPlayer()

	fmt.Println("[Test.Player]: exists -> 76561198074578368")

	player.SetSteamId("76561198074578368") // some steamrep admin
	err, exists := player.Exists()

	if exists {
		fmt.Println("[Test.Player.Exists]: player exists!")
	} else {
		fmt.Println("[Test.Player.Exists]: can't find player!")
	}

	if err != nil {
		log.Fatal(err)
	}
}

// test if can find a player
func TestDatabaseFind(t *testing.T) {
	player := models.NewPlayer()

	fmt.Println("[Test.Player]: finding -> 76561198074578368")

	// sets player's steamId then try to find it
	player.SetSteamId("76561198074578368") // some steamrep admin
	err := player.Find()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("[Test.Player.Find]: id -> " + player.GetId())
		fmt.Println("[Test.Player.Find]: name -> " + player.Name)
		fmt.Println("[Test.Player.Find]: steamid -> " + player.SteamId)
		fmt.Println("[Test.Player.Find]: created -> " + player.GetCreated())
	}
}

// test -> remove player
func TestDatabaseDelete(t *testing.T) {
	player := models.NewPlayer()

	fmt.Println("[Test.Player.Delete]: Deleting -> 76561198074578368")

	player.SetSteamId("76561198074578368") // some steamrep admin
	err := player.Delete()

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("[Test.Player.Delete]: deleted player!")
	}
}

// test if the player doesnt exists
func TestDatabaseDoesntExists(t *testing.T) {
	player := models.NewPlayer()

	fmt.Println("[Test.Player.Doesnt_Exists]: exists -> 76561198074578368")

	player.SetSteamId("76561198074578368") // some steamrep admin
	err, exists := player.Exists()

	if exists {
		fmt.Println("[Test.Player.Doesnt_Exists]: player exists!")
	} else {
		fmt.Println("[Test.Player.Doesnt_Exists]: can't find player!")
	}

	if err == nil {
		log.Fatal(err)
	}
}

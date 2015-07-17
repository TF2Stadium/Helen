package database

import (
	"fmt"
	"log"
	"time"

	"github.com/TF2Stadium/Server/config"
	"gopkg.in/mgo.v2"
)

/**
 * Stores the connection with mongodb
 */
var Connection *mgo.Session
var Database *mgo.Database

// we'll use Test() to set this
// will only use to change main db name
var IsTest bool = false

// we'll connect to the database through this function
func Init() {
	fmt.Println("[Database]: Database name -> [" + getDatabaseName() + "]")
	fmt.Println("[Database]: Database user -> [" + config.Constants.DbUsername + "]")
	fmt.Println("[Database]: Connecting to database -> [" + config.Constants.DbDatabase + "]")

	// mdb connection
	info := &mgo.DialInfo{
		Addrs:    []string{config.Constants.DbHosts},
		Timeout:  60 * time.Second,
		Database: getDatabaseName(),
		Username: config.Constants.DbUsername,
		Password: config.Constants.DbPassword,
	}

	// starts a connection with the mongodb server
	con, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Fatal(err)
	}

	con.SetMode(mgo.Strong, true)

	Connection = con
	Database = con.DB(getDatabaseName())

	fmt.Println("[Database]: Connected!")
}

func GetCollection(collection string) *mgo.Collection {
	return Connection.DB(getDatabaseName()).C(collection)
}

func GetLobbiesCollection() *mgo.Collection {
	return GetCollection(config.Constants.DbLobbiesCollection)
}
func GetPlayersCollection() *mgo.Collection {
	return GetCollection(config.Constants.DbPlayersCollection)
}

func Get(collection string) (*mgo.Session, *mgo.Collection) {
	log.Println("[Database.Session]: Getting new session...")

	// creates a new session
	sess := Connection.Copy()

	log.Println("[Database.Session]: Getting collection -> [" + collection + "]")

	// gets the specified collection
	col := sess.DB(getDatabaseName()).C(collection)

	log.Println("[Database.Session]: Got session, returning!")
	return sess, col
}

func Test() {
	IsTest = true
}

func getDatabaseName() string {
	if IsTest {
		return config.Constants.DbTestDatabase
	}

	return config.Constants.DbDatabase
}

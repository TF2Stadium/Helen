package database

import (
	"fmt"
	"github.com/TeamPlayTF/Server/config"
	"gopkg.in/mgo.v2"
	"log"
	"time"
)

/**
 * Stores the connection with mongodb
 */
var Conection *mgo.Session
var Database *mgo.Database

// we'll use Test() to set this
// will only use to change main db name
var IsTest bool = false

// we'll connect to the database through this function
func Init() {
	fmt.Println("[Database]: Database name -> [" + getDatabaseName() + "]")
	fmt.Println("[Database]: Database user -> [" + config.Constants.DbUsername + "]")
	fmt.Println("[Database]: Database pass -> [" + config.Constants.DbPassword + "]")
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

	con.SetMode(mgo.Monotonic, true)

	Conection = con
	Database = con.DB(getDatabaseName())

	fmt.Println("[Database]: Connected!")
}

func Get(collection string) (*mgo.Session, *mgo.Collection) {
	fmt.Println("[Database.Session]: Getting new session...")

	// creates a new session
	sess := Conection.Copy()

	fmt.Println("[Database.Session]: Getting collection -> [" + collection + "]")

	// gets the specified collection
	col := sess.DB(getDatabaseName()).C(collection)

	fmt.Println("[Database.Session]: Got session, returning!")
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

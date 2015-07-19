package database

import (
	"fmt"
	"log"

	"github.com/TF2Stadium/Server/config"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// we'll use Test() to set this
// will only use to change main db name
var IsTest bool = false

var DB gorm.DB
var DbUrl string

// we'll connect to the database through this function
func Init() {

	fmt.Println("[DB]: DB name -> [" + getDatabaseName() + "]")
	fmt.Println("[DB]: DB user -> [" + config.Constants.DbUsername + "]")
	fmt.Println("[DB]: Connecting to database -> [" + config.Constants.DbDatabase + "]")

	DbUrl = "postgres://" + config.Constants.DbUsername + ":" +
		config.Constants.DbPassword + "@" +
		config.Constants.DbHost + "/" +
		getDatabaseName() + "?sslmode=disable"

	var err error
	DB, err = gorm.Open("postgres", DbUrl)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("[DB]: Connected!")
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

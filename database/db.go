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
	fmt.Println("[DB]: DB name -> [" + config.Constants.DbDatabase + "]")
	fmt.Println("[DB]: DB user -> [" + config.Constants.DbUsername + "]")
	fmt.Println("[DB]: Connecting to database -> [" + config.Constants.DbDatabase + "]")

	var passwordArg string
	if config.Constants.DbPassword == "" {
		passwordArg = ""
	} else {
		passwordArg = ":" + config.Constants.DbPassword
	}

	DbUrl = "postgres://" + config.Constants.DbUsername +
		passwordArg + "@" +
		config.Constants.DbHost + "/" +
		config.Constants.DbDatabase + "?sslmode=disable"

	var err error
	DB, err = gorm.Open("postgres", DbUrl)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("[DB]: Connected!")
}

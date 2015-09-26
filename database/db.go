// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package database

import (
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"sync"
)

// we'll use Test() to set this
// will only use to change main db name
var IsTest bool = false

var DB gorm.DB
var dbMutex sync.Mutex
var initialized = false
var DbUrl string

// we'll connect to the database through this function
func Init() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if initialized {
		return
	}

	helpers.Logger.Debug("[DB]: DB name -> [" + config.Constants.DbDatabase + "]")
	helpers.Logger.Debug("[DB]: DB user -> [" + config.Constants.DbUsername + "]")
	helpers.Logger.Debug("[DB]: Connecting to database -> [" + config.Constants.DbDatabase + "]")

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
	//	DB, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		helpers.Logger.Fatal(err.Error())
	}

	DB.SetLogger(helpers.FakeLogger{})

	helpers.Logger.Debug("[DB]: Connected!")
	initialized = true
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package database

import (
	"net/url"
	"sync"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// we'll use Test() to set this
// will only use to change main db name
var IsTest bool = false

var DB gorm.DB
var dbMutex sync.Mutex
var initialized = false
var DBUrl url.URL

// we'll connect to the database through this function
func Init() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if initialized {
		return
	}

	DBUrl = url.URL{
		Scheme:   "postgres",
		Host:     config.Constants.DbHost,
		Path:     config.Constants.DbDatabase,
		RawQuery: "sslmode=disable",
	}

	helpers.Logger.Info("Connecting to DB on %s", DBUrl.String())

	DBUrl.User = url.UserPassword(config.Constants.DbUsername, config.Constants.DbPassword)

	var err error
	DB, err = gorm.Open("postgres", DBUrl.String())
	//	DB, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		helpers.Logger.Fatal(err.Error())
	}

	DB.SetLogger(helpers.FakeLogger{})

	helpers.Logger.Info("Connected!")
	initialized = true
}

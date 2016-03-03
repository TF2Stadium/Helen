// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package database

import (
	"flag"
	"net/url"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// we'll use Test() to set this
// will only use to change main db name
var (
	IsTest      bool = false
	DB          gorm.DB
	dbMutex     sync.Mutex
	initialized = false
	DBUrl       url.URL
	maxOpen     = flag.Int("maxopen", 100, "maximum number of open database connections")
)

// we'll connect to the database through this function
func Init() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if initialized {
		return
	}

	DBUrl = url.URL{
		Scheme:   "postgres",
		Host:     config.Constants.DbAddr,
		Path:     config.Constants.DbDatabase,
		RawQuery: "sslmode=disable",
	}

	logrus.Info("Connecting to DB on ", DBUrl.String())

	DBUrl.User = url.UserPassword(config.Constants.DbUsername, config.Constants.DbPassword)

	var err error
	DB, err = gorm.Open("postgres", DBUrl.String())
	//	DB, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		logrus.Fatal(err.Error())
	}

	DB.DB().SetMaxOpenConns(*maxOpen)
	DB.SetLogger(logrus.StandardLogger())

	logrus.Info("Connected!")
	initialized = true
}

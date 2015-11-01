// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
)

var roomFile = make(map[uint]*os.File)

func CheckLogger() {
	path, err := filepath.Abs(config.Constants.ChatLogsDir)
	if err != nil {
		helpers.Logger.Fatalf("%s", err.Error())
	}

	os.Mkdir(path, 0666)
}

func LogChat(room uint, name string, steamid string, message string) {
	if !config.Constants.ChatLogsEnabled {
		return
	}

	year, month, day := time.Now().Date()
	path, err := filepath.Abs(config.Constants.ChatLogsDir)
	if err != nil {
		helpers.Logger.Fatalf("%s", err.Error())
	}

	fileName := fmt.Sprintf("%s/room%d_%d_%d_%d", path, room, day, month, year)

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		helpers.Logger.Fatalf("%s", err.Error())
	}

	helpers.Logger.Debug("%s: %s", name, message)
	file.Seek(0, 2)
	fmt.Fprintf(file, "%s<%s>: %s\n", name, steamid, message)
	file.Close()
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/vibhavp/wsevent"
)

type Param struct {
	Kind    reflect.Kind
	Default interface{}
	In      interface{}
}

type FilterParams struct {
	Action      authority.AuthAction
	FilterLogin bool
	Params      map[string]Param
}

var WhitelistSteamID = make(map[string]bool)

func InitSteamIDWhitelist(filename string) {
	absName, _ := filepath.Abs(filename)
	data, _ := ioutil.ReadFile(absName)
	ids := strings.Split(string(data), "\n")

	for _, id := range ids {
		helpers.Logger.Debug("Whitelisting SteamID %s", id)
		WhitelistSteamID[id] = true
	}
}

func FilterRequest(so *wsevent.Client, action authority.AuthAction, login bool) (err *helpers.TPError) {
	if int(action) != 0 {
		var role, _ = GetPlayerRole(so.Id())
		can := role.Can(action)
		if !can {
			err = helpers.NewTPError("You are not authorized to perform this action.", 0)
		}
	}
	return
}

func GetParams(data string, i interface{}) error {
	return json.Unmarshal([]byte(data), i)
}

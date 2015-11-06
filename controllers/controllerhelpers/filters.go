// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/vibhavp/wsevent"
)

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
	if login && !IsLoggedInSocket(so.Id()) {
		return helpers.NewTPError("Player isn't logged in.", -4)

	}
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
	err := json.Unmarshal([]byte(data), i)

	if err != nil {
		return err
	}

	stValue := reflect.Indirect(reflect.ValueOf(i))
	stType := stValue.Type()

outer:
	for i := 0; i < stType.NumField(); i++ {
		field := stType.Field(i)
		fieldValue := stValue.Field(i)
		if field.Type.Kind() != reflect.String {
			if fieldValue.IsNil() {
				return errors.New(fmt.Sprintf(`Field "%s" cannot be null`,
					strings.ToLower(field.Name)))
			}
		} else if fieldValue.String() == "" {
			return errors.New(fmt.Sprintf(`Field "%s" cannot be null`,
				strings.ToLower(field.Name)))

		}

		validTag := field.Tag.Get("valid")
		if validTag == "" {
			continue
		}

		arr := strings.Split(validTag, ",")
		var valid bool
		for _, validVal := range arr {
			switch field.Type.Kind() {
			case reflect.Uint:
				num, err := strconv.ParseUint(validVal, 2, 32)
				if err != nil {
					panic(err.Error())
				}

				if reflect.DeepEqual(fieldValue.Uint(), num) {
					valid = true
					continue outer
				}

			case reflect.String:
				if reflect.DeepEqual(fieldValue.String(), validVal) {
					valid = true
					continue outer
				}

			}
		}
		if !valid {
			return errors.New(fmt.Sprintf("Field %s isn't valid.", field.Name))
		}
	}

	return nil
}

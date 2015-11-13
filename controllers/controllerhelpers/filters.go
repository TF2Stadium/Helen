// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/wsevent"
)

var whitelistLock = new(sync.RWMutex)
var whitelistSteamID map[string]bool

func WhitelistListener() {
	ticker := time.NewTicker(time.Minute * 5)
	for {
		resp, err := http.Get(config.Constants.SteamIDWhitelist)
		if err != nil {
			continue
		}

		bytes, _ := ioutil.ReadAll(resp.Body)
		var groupXML struct {
			//XMLName xml.Name `xml:"memberList"`
			//GroupID uint64   `xml:"groupID64"`
			Members []string `xml:"members>steamID64"`
		}

		xml.Unmarshal(bytes, &groupXML)

		whitelistLock.Lock()
		whitelistSteamID = make(map[string]bool)

		for _, steamID := range groupXML.Members {
			//_, ok := whitelistSteamID[steamID]
			helpers.Logger.Info("Whitelisting SteamID %s", steamID)
			whitelistSteamID[steamID] = true
		}
		whitelistLock.Unlock()
		<-ticker.C
	}
}

func IsSteamIDWhitelisted(steamid string) bool {
	whitelistLock.RLock()
	defer whitelistLock.RUnlock()
	whitelisted, exists := whitelistSteamID[steamid]

	helpers.Logger.Debug("steamid: %s, whitelisted: %t, eixsts: %t", steamid, whitelisted, exists)

	return whitelisted && exists
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

func GetParams(data string, v interface{}) error {
	err := json.Unmarshal([]byte(data), v)

	if err != nil {
		return err
	}

	stValue := reflect.Indirect(reflect.ValueOf(v))
	stType := stValue.Type()

	for i := 0; i < stType.NumField(); i++ {
		field := stType.Field(i)
		fieldPtrValue := stValue.Field(i)             //The pointer field
		fieldValue := reflect.Indirect(fieldPtrValue) //The value to which the pointer points too

		if field.Type.Kind() != reflect.String {
			if fieldPtrValue.IsNil() {
				emptyTag := field.Tag.Get("empty")
				if emptyTag == "" {
					return errors.New(fmt.Sprintf(`Field "%s" cannot be null`,
						strings.ToLower(field.Name)))
				}

				switch fieldPtrValue.Type().Elem().Kind() {
				case reflect.Uint:
					num, err := strconv.ParseUint(emptyTag, 2, 32)
					if err != nil {
						panic(err.Error())
					}
					fieldPtrValue.Set(reflect.ValueOf(&num))
				case reflect.String:
					fieldPtrValue.Set(reflect.ValueOf(&emptyTag))
				case reflect.Bool:
					b, ok := map[string]bool{
						"true":  true,
						"false": false}[emptyTag]
					if !ok {
						panic(fmt.Sprintf(
							"%s is not a valid boolean literal string",
							emptyTag))
					}
					fieldPtrValue.Set(reflect.ValueOf(&b))
				}
				continue
			}
		} else if fieldValue.String() == "" && field.Tag.Get("empty") == "" {
			return errors.New(fmt.Sprintf(`Field "%s" cannot be null`,
				strings.ToLower(field.Name)))
		}

		validTag := field.Tag.Get("valid")
		if validTag == "" {
			continue
		}

		arr := strings.Split(validTag, ",")
		var valid bool
	outer:
		for _, validVal := range arr {
			switch fieldValue.Kind() {
			case reflect.Uint:
				num, err := strconv.ParseUint(validVal, 2, 32)
				if err != nil {
					panic(fmt.Sprintf("Error while parsing struct tag: %s",
						err.Error()))
				}

				if reflect.DeepEqual(fieldValue.Uint(), num) {
					valid = true
					break outer
				}

			case reflect.String:
				if reflect.DeepEqual(fieldValue.String(), validVal) {
					valid = true
					break outer
				}

			}
		}
		if !valid {
			return errors.New(fmt.Sprintf("Field %s isn't valid.", field.Name))
		}
	}

	return nil
}

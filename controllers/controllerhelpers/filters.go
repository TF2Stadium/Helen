// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"encoding/json"
	"encoding/xml"
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
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

var (
	whitelistLock    = new(sync.RWMutex)
	whitelistSteamID map[string]bool
)

func WhitelistListener() {
	ticker := time.NewTicker(time.Minute * 1)
	for {
		resp, err := http.Get(config.Constants.SteamIDWhitelist)

		if err != nil {
			helpers.Logger.Error(err.Error())
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
			//helpers.Logger.Info("Whitelisting SteamID %s", steamID)
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

	return whitelisted && exists
}

// shitlord
func CheckPrivilege(so *wsevent.Client, action authority.AuthAction) (err *helpers.TPError) {
	//Checks if the client has the neccesary authority to perform action
	if int(action) != 0 {
		var role, _ = GetPlayerRole(so.Id())
		can := role.Can(action)
		if !can {
			err = helpers.NewTPError("You are not authorized to perform this action.", 0)
		}
	}
	return
}

func FilterHTTPRequest(action authority.AuthAction, f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSessionHTTP(r)
		if err != nil {
			http.Error(w, "Internal Server Error: No session found", 500)
			return
		}

		steamid, ok := session.Values["steam_id"]
		if !ok {
			http.Error(w, "Player not logged in", 401)
			return
		}

		player, _ := models.GetPlayerBySteamID(steamid.(string))
		if !(player.Role.Can(action)) {
			http.Error(w, "Not authorized", 403)
			return
		}

		f(w, r)
	}
}

//I forgot to document this while working on it, so it might be a bit
//difficult to understand what's going on.
//THINK TWICE BEFORE CHANGING ANYTHING HERE
func GetParams(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)

	if err != nil {
		return err
	}

	stValue := reflect.Indirect(reflect.ValueOf(v))
	stType := stValue.Type()

outer:
	for i := 0; i < stType.NumField(); i++ {
		field := stType.Field(i)
		fieldPtrValue := stValue.Field(i) //The pointer field

		if fieldPtrValue.Type().Elem().Kind() != reflect.String {
			if fieldPtrValue.IsNil() && field.Tag.Get("empty") == "" {
				return fmt.Errorf(`Field "%s" cannot be null`, strings.ToLower(field.Name))
			}
		} else if fieldPtrValue.IsNil() {
			empty := field.Tag.Get("empty")
			if empty == "-" {
				empty = ""
			} else {
				return fmt.Errorf(`Field "%s" cannot be null`, strings.ToLower(field.Name))
			}
			fieldPtrValue.Set(reflect.ValueOf(&empty))
		}

		validTag := field.Tag.Get("valid")
		if validTag == "" {
			continue
		}

		fieldValue := reflect.Indirect(fieldPtrValue) //The value to which the pointer points too
		validValues := strings.Split(validTag, ",")
		var valid bool

		for _, validVal := range validValues {
			switch fieldValue.Kind() {
			case reflect.Uint:
				num, err := strconv.ParseUint(validVal, 2, 32)
				if err != nil {
					panic(fmt.Errorf("Error while parsing struct tag: %s",
						err.Error()))
				}

				if reflect.DeepEqual(fieldValue.Uint(), num) {
					continue outer
				}

			case reflect.String:
				if reflect.DeepEqual(fieldValue.String(), validVal) {
					continue outer
				}

			}
		}
		if !valid {
			return fmt.Errorf("Field %s isn't valid.", field.Name)
		}
	}

	return nil
}

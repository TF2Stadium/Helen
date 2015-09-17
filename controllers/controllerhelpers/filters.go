// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
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

func FilterRequest(so socketio.Socket, filters FilterParams, f func(map[string]interface{}) string) func(string) string {

	return func(jsonStr string) string {
		if filters.FilterLogin && !IsLoggedInSocket(so.Id()) {

			bytes, _ := BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

		// Careful: this assumes normal players can do everything (since helpers.RolePlayer==0)
		if int(filters.Action) != 0 {
			var role, _ = GetPlayerRole(so.Id())
			can := role.Can(filters.Action)
			if !can {
				bytes, _ := BuildFailureJSON("You are not authorized to perform this action.", 0).Encode()
				return string(bytes)
			}
		}

		if filters.Params == nil {
			return f(nil)
		}

		js, err := simplejson.NewFromReader(strings.NewReader(jsonStr))
		if err != nil {
			bytes, _ := BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

		paramMap, err := js.Map()
		if err != nil {
			bytes, _ := BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

	outer:
		for key, param := range filters.Params {
			_, ok := paramMap[key]

			if !ok {
				if param.Default == nil {
					bytes, _ := BuildMissingArgJSON(key).Encode()
					return string(bytes)
				}
				paramMap[key] = param.Default
			}

			if kind := reflect.ValueOf(paramMap[key]).Kind(); kind != param.Kind {
				if param.Kind == reflect.Uint {
					if num, err := js.Get(key).Uint64(); err == nil {
						paramMap[key] = uint(num)
						continue
					}
				} else if param.Kind == reflect.Int {
					if num, err := js.Get(key).Int64(); err == nil {
						paramMap[key] = int(num)
						continue
					}
				}
				bytes, _ := BuildMissingArgJSON(key).Encode()
				return string(bytes)
			}

			errFormat := `Paramter "%s" not valid`

			if param.In != nil {
				switch param.Kind {
				case reflect.String:
					for _, val := range param.In.([]string) {
						if paramMap[key] == val {
							continue outer
						}
					}

				case reflect.Int:
					for _, val := range param.In.([]int) {
						if paramMap[key] == val {
							continue outer
						}
					}

				case reflect.Uint:
					for _, val := range param.In.([]uint) {
						if paramMap[key] == val {
							continue outer
						}
					}
				}
				bytes, _ := BuildFailureJSON(fmt.Sprintf(errFormat, key), 0).Encode()
				return string(bytes)
			}
		}

		return f(paramMap)
	}
}

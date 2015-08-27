package controllerhelpers

import (
	"reflect"
	"strings"

	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
)

type Param struct {
	Kind    reflect.Kind
	Default interface{}
	In      []string
}

func RegisterEvent(so socketio.Socket, event string, params map[string]Param,
	action authority.AuthAction, f func(map[string]interface{}) string) {

	so.On(event, func(jsonStr string) string {
		if !IsLoggedInSocket(so.Id()) {
			bytes, _ := BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

		if int(action) != 0 {
			var role, _ = GetPlayerRole(so.Id())
			can := role.Can(action)
			if !can {
				bytes, _ := BuildFailureJSON("You are not authorized to perform this action.", 0).Encode()
				return string(bytes)
			}
		}

		if params == nil {
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

		for key, param := range params {
			value, ok := paramMap[key]
			if !ok {
				if param.Default == nil {
					bytes, _ := BuildMissingArgJSON(key).Encode()
					return string(bytes)
				}
				paramMap[key] = param.Default
			}

			if kind := reflect.ValueOf(value).Kind(); kind != param.Kind {
				if param.Kind == reflect.Uint {
					if num, err := js.Get(key).Uint64(); err == nil {
						paramMap[key] = uint(num)
						continue
					}
				} else if param.Kind == reflect.Int {
					if num, err := js.Get(key).Int64(); err == nil {
						paramMap[key] = int64(num)
						continue
					}
				}
				bytes, _ := BuildMissingArgJSON(key).Encode()
				return string(bytes)
			}
		}
		return f(paramMap)
	})
}

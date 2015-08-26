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

		var role, _ = GetPlayerRole(so.Id())
		can := role.Can(action)
		if !can {
			bytes, _ := BuildFailureJSON("You are not authorized to perform this action.", 0).Encode()
			return string(bytes)
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
					switch kind {
					case reflect.Uint, reflect.Uint64, reflect.Int:
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

package controllerhelpers

import (
	"strings"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/bitly/go-simplejson"
)

func AuthFilter(socketid string, f func(string) string) func(string) string {
	return func(data string) string {
		if !IsLoggedInSocket(socketid) {
			bytes, _ := BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}
		return f(data)
	}
}

func JsonParamFilter(f func(*simplejson.Json) string) func(string) string {
	return func(data string) string {
		js, err := simplejson.NewFromReader(strings.NewReader(data))
		if err != nil {
			bytes, _ := BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

		return f(js)
	}
}

func AuthorizationFilter(socketid string, action authority.AuthAction, f func(string) string) func(string) string {
	return AuthFilter(socketid, func(data string) string {
		var role, _ = GetPlayerRole(socketid)
		can := role.Can(action)
		if !can {
			bytes, _ := BuildFailureJSON("You are not authorized to perform this action.", 0).Encode()
			return string(bytes)
		}

		return f(data)
	})
}

type Param struct {
	Type    ParamType
	Default interface{}
}

type ParamType int

const (
	PTypeInt    ParamType = iota
	PTypeString ParamType = iota
	PTypeBool   ParamType = iota
	PTypeFloat  ParamType = iota
)

func JsonVerifiedFilter(p map[string]Param, f func(*simplejson.Json) string) func(string) string {
	return JsonParamFilter(func(js *simplejson.Json) string {
		for name, paramtype := range p {
			switch paramtype.Type {
			case PTypeInt:
				_, err := js.Get(name).Int()
				if err != nil && paramtype.Default == nil {
					bytes, _ := BuildMissingArgJSON(name).Encode()
					return string(bytes)
				} else if err != nil {
					js.Set(name, paramtype.Default.(int))
				}
			case PTypeString:
				_, err := js.Get(name).String()
				if err != nil && paramtype.Default == nil {
					bytes, _ := BuildMissingArgJSON(name).Encode()
					return string(bytes)
				} else if err != nil {
					js.Set(name, paramtype.Default.(string))
				}
			case PTypeBool:
				_, err := js.Get(name).Bool()
				if err != nil && paramtype.Default == nil {
					bytes, _ := BuildMissingArgJSON(name).Encode()
					return string(bytes)
				} else if err != nil {
					js.Set(name, paramtype.Default.(bool))
				}
			case PTypeFloat:
				_, err := js.Get(name).Float64()
				if err != nil && paramtype.Default == nil {
					bytes, _ := BuildMissingArgJSON(name).Encode()
					return string(bytes)
				} else if err != nil {
					js.Set(name, paramtype.Default.(float64))
				}
			default:
				helpers.Logger.Panicf("Invalid type as parameter type for %s: %d", name, paramtype)
			}
		}

		return f(js)
	})
}

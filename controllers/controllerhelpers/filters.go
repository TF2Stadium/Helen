package controllerhelpers

import (
	"strings"

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

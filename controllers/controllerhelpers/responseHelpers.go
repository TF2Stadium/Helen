// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"github.com/bitly/go-simplejson"
)

func BuildSuccessJSON(data interface{}) *simplejson.Json {
	j := simplejson.New()
	j.Set("success", true)
	j.Set("data", data)

	return j
}

var emptyBytes, _ = BuildSuccessJSON(simplejson.New()).Encode()
var EmptySuccessJS = string(emptyBytes)

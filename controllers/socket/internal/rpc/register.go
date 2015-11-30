// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package rpc

import (
	"reflect"

	"github.com/TF2Stadium/wsevent"
)

func Register(server *wsevent.Server, rcvr interface{}) {
	rval := reflect.ValueOf(rcvr)
	rtype := reflect.TypeOf(rcvr)

	for i := 0; i < rval.NumMethod(); i++ {
		method := rval.Method(i)
		name := rtype.Method(i).Name
		name = string((name[0])+32) + name[1:]

		server.On(name, func(s *wsevent.Server, c *wsevent.Client, b []byte) []byte {
			rtrn := method.Call([]reflect.Value{
				reflect.ValueOf(s), reflect.ValueOf(c), reflect.ValueOf(b)})
			return rtrn[0].Bytes()
		})

	}
}

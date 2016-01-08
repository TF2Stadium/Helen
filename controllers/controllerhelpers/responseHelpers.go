// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import "encoding/json"

//Response stores a successful response to a RPC call
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func NewResponse(data interface{}) Response {
	return Response{true, data}
}

func (r Response) Encode() ([]byte, error) {
	return json.Marshal(r)
}

var EmptySuccessJS = NewResponse(struct{}{})

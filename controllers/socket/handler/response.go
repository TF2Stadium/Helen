// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

//Response stores a successful response to a RPC call
type response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

//Create a new response
func newResponse(data interface{}) response {
	return response{true, data}
}

//EmptySuccessJS is the empty success response
var emptySuccess = newResponse(struct{}{})

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"encoding/json"
)

type TPError struct {
	Str  string `json:"message"`
	Code int    `json:"code"`
	//For the json object
	Success bool `json:"success"`
}

func (e *TPError) Error() string {
	return e.Str
}

func NewTPError(str string, code int) *TPError {
	return &TPError{
		Str:  str,
		Code: code}
}

func NewTPErrorFromError(e error) *TPError {
	if e == nil {
		return nil
	}
	return &TPError{
		Str:  e.Error(),
		Code: 0,
	}
}

func (e *TPError) Encode() ([]byte, error) {
	return json.Marshal(e)
}

package helpers

import "github.com/bitly/go-simplejson"

type TPError struct {
	Str  string
	Code int
}

func (e *TPError) Error() string {
	return e.Str
}

func NewTPError(str string, code int) *TPError {
	return &TPError{
		Str:  str,
		Code: code}
}

func (e *TPError) ErrorJSON() *simplejson.Json {
	j := simplejson.New()

	j.Set("success", false)
	j.Set("message", e.Str)
	j.Set("code", e.Code)

	return j
}

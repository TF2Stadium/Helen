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

func NewTPErrorFromError(e error) *TPError {
	if e == nil {
		return nil
	}
	return &TPError{
		Str:  e.Error(),
		Code: 0,
	}
}

func (e *TPError) ErrorJSON() *simplejson.Json {
	j := simplejson.New()

	j.Set("success", false)
	j.Set("message", e.Str)
	j.Set("code", e.Code)

	return j
}

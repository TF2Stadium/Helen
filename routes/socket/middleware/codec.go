package middleware

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type JSONCodec struct{}

func (JSONCodec) ReadName(data []byte) string {
	var body struct {
		Request string
	}
	json.Unmarshal(data, &body)
	return body.Request
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)

	if err != nil {
		return err
	}

	stValue := reflect.Indirect(reflect.ValueOf(v))

outer:
	for i := 0; i < stValue.NumField(); i++ {
		curField := stValue.Type().Field(i)
		fieldPtrValue := stValue.Field(i) //The pointer field

		if fieldPtrValue.Kind() != reflect.Ptr { //field isn't a pointer, continue
			continue
		}

		if fieldPtrValue.IsNil() {
			if curField.Tag.Get("empty") == "" {
				return fmt.Errorf(`Field "%s" cannot be null`, strings.ToLower(curField.Name))
			}

			if fieldPtrValue.Type().Elem().Kind() == reflect.String {
				blank := ""
				fieldPtrValue.Set(reflect.ValueOf(&blank))
			}
		}

		//If a list of valid strings is supplied, check if the field
		//is valid
		validTag := curField.Tag.Get("valid")
		if validTag == "" {
			continue
		}

		for _, valid := range strings.Split(validTag, ",") {
			if reflect.DeepEqual(reflect.Indirect(fieldPtrValue).String(), valid) {
				continue outer
			}
		}
		return fmt.Errorf("Field %s isn't valid.", curField.Name)
	}

	return nil
}

func (JSONCodec) Error(err error) interface{} {
	return struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
	}{err.Error(), false}
}

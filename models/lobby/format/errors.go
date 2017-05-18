package format

import (
	"strconv"
)

type ErrorInvalidTeam string

func (e ErrorInvalidTeam) Error() string {
	return "format: Invalid Team: " + string(e)
}

type ErrorInvalidClass string

func (e ErrorInvalidClass) Error() string {
	return "format: Invalid Class: " + string(e)
}

type ErrorInvalidSlot int

func (e ErrorInvalidSlot) Error() string {
	return "format: Invalid Slot number: " + strconv.Itoa(int(e))
}

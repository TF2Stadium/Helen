package helpers

import (
	"testing"
)

func TestNewError(t *testing.T) {
	t.Log(NewError("Hey this is an error", 1))
}

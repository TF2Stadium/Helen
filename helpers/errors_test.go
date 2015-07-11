package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	err := NewTPError("Hey this is an error", 1)

	assert.Equal(t, err.Error(), "Hey this is an error")
}

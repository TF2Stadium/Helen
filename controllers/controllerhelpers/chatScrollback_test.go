package controllerhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddScrollbackMessage(t *testing.T) {
	InitChatScrollback()
	AddScrollbackMessage("foo")
	AddScrollbackMessage("bar")
	AddScrollbackMessage("baz")

	c := chatScrollback.first
	assert.Equal(t, c.Value.(string), "foo")
	c = c.Next()
	assert.Equal(t, c.Value.(string), "bar")
	c = c.Next()
	assert.Equal(t, c.Value.(string), "baz")
}

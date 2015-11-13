package controllerhelpers

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddScrollbackMessage(t *testing.T) {
	messages := make([]string, 20)
	for i, _ := range messages {
		messages[i] = strconv.Itoa(i)
		AddScrollbackMessage(0, messages[i])
	}

	c := chatScrollback[0].first
	for _, message := range messages {
		assert.Equal(t, message, c.Value.(string))
		c = c.Next()
	}
}

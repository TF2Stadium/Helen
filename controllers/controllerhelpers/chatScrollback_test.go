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
		AddScrollbackMessage(1, messages[i])
	}

	curr := chatScrollback[0].curr
	c := chatScrollback[0]
	c2 := chatScrollback[1]

	for _, _ = range messages {
		assert.Equal(t, c2.messages[curr], c.messages[curr])

		curr += 1
		if curr == 20 {
			curr = 0
		}
	}
}

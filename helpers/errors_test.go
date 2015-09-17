// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	InitLogger()
}

func TestNewError(t *testing.T) {
	err := NewTPError("Hey this is an error", 1)

	assert.Equal(t, err.Error(), "Hey this is an error")
}

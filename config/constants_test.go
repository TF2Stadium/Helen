// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package config_test

import (
	"os"
	"testing"

	. "github.com/TF2Stadium/Helen/config"
	_ "github.com/TF2Stadium/Helen/helpers"
	"github.com/stretchr/testify/assert"
)

func TestEnvVariablesOverrideConfig(t *testing.T) {
	os.Unsetenv("SERVER_ADDR")
	SetupConstants()
	addr := Constants.ListenAddress

	os.Setenv("SERVER_ADDR", "123456as")
	SetupConstants()
	addr2 := Constants.ListenAddress

	assert.NotEqual(t, addr, addr2)
}

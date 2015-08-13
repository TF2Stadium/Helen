package config

import (
	"os"
	"testing"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/stretchr/testify/assert"
)

func init() {
	helpers.InitLogger()
}

func TestEnvVariablesOverrideConfig(t *testing.T) {
	os.Unsetenv("PORT")
	SetupConstants()
	port := Constants.Port

	os.Setenv("PORT", "123456as")
	SetupConstants()
	port2 := Constants.Port

	assert.NotEqual(t, port, port2)
}

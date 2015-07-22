package database

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/TF2Stadium/Server/config"
	"github.com/stretchr/testify/assert"
)

var steamid = "76561198074578368"

func TestDatabasePing(t *testing.T) {
	if os.Getenv("DEPLOYMENT_ENV") == "" {
		os.Setenv("DEPLOYMENT_ENV", "test")
		defer os.Unsetenv("DEPLOYMENT_ENV")
	}
	config.SetupConstants()

	fmt.Println("[Test.Database] IsTest? " + strconv.FormatBool(IsTest))
	Init()
	assert.Nil(t, DB.DB().Ping())
}

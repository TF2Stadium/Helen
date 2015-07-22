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
	os.Setenv("DEPLOYMENT_ENV", "test")
	config.SetupConstants()
	os.Unsetenv("DEPLOYMENT_ENV")

	fmt.Println("[Test.Database] IsTest? " + strconv.FormatBool(IsTest))
	Init()
	assert.Nil(t, DB.DB().Ping())
}

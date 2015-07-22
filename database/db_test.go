package database

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/TF2Stadium/Server/config"
	"github.com/stretchr/testify/assert"
)

var steamid = "76561198074578368"

func TestDatabasePing(t *testing.T) {
	config.SetupConstants()
	Test()
	fmt.Println("[Test.Database] IsTest? " + strconv.FormatBool(IsTest))
	Init()
	assert.Nil(t, DB.DB().Ping())
}

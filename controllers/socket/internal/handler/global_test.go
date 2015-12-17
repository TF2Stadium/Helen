package handler_test

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/TF2Stadium/Helen/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestSocketInfo(t *testing.T) {
	testhelpers.CleanupDB()
	server := testhelpers.StartServer(testhelpers.NewSockets())
	defer server.Close()

	client := testhelpers.NewClient()
	testhelpers.Login(strconv.Itoa(rand.Int()), client)
	conn, err := testhelpers.ConnectWS(client)
	defer conn.Close()
	assert.NoError(t, err)

	_, err = testhelpers.ReadMessages(conn, testhelpers.InitMessages, nil)
	assert.NoError(t, err)

	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "getSocketInfo",
		},
	}

	reply, err := testhelpers.EmitJSONWithReply(conn, args)
	assert.NoError(t, err)
	data := reply["data"].(map[string]interface{})
	assert.Equal(t, data["rooms"].([]interface{})[0].(string), "0_public")
	//t.Logf("%v", reply)

}

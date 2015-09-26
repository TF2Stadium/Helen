package socket

import (
	"github.com/TF2Stadium/Helen/testhelpers"
	"testing"
	//	"github.com/TF2Stadium/Helen/helpers"
	"github.com/stretchr/testify/assert"
)

func TestPlayerSettings(t *testing.T) {
	testhelpers.CleanupDB()

	player := testhelpers.CreatePlayer()

	pSocket := testhelpers.NewFakeSocket()
	pSocket.FakeAuthenticate(player)

	SocketInit(pSocket)

	// should return empty response if there are no settings
	res, _ := pSocket.SimRequest("playerSettingsGet", "{}")
	d := testhelpers.UnpackSuccessResponse(t, res)
	assert.Equal(t, 0, len(d.MustMap()))

	// should return empty response if key not set
	res2, _ := pSocket.SimRequest("playerSettingsGet", `{"key": "apples"}`)
	testhelpers.UnpackFailureResponse(t, res2)

	// should set value successfully
	res3, _ := pSocket.SimRequest("playerSettingsSet", `{"key": "apples", "value": "delicious"}`)
	testhelpers.UnpackSuccessResponse(t, res3)

	res4, _ := pSocket.SimRequest("playerSettingsSet", `{"key": "pears", "value": "disguisting"}`)
	testhelpers.UnpackSuccessResponse(t, res4)

	// should only set strings
	res5, _ := pSocket.SimRequest("playerSettingsSet", `{"key": "apples", "value": true}`)
	testhelpers.UnpackFailureResponse(t, res5)

	res6, _ := pSocket.SimRequest("playerSettingsSet", `{"key": "apples", "value": [1,2,3]}`)
	testhelpers.UnpackFailureResponse(t, res6)

	// should return existing values
	res7, _ := pSocket.SimRequest("playerSettingsGet", `{"key": "apples"}`)
	d7 := testhelpers.UnpackSuccessResponse(t, res7)
	assert.Equal(t, "delicious", d7.Get("apples").MustString())

	// should return all values if key not specified
	res8, _ := pSocket.SimRequest("playerSettingsGet", `{}`)
	d8 := testhelpers.UnpackSuccessResponse(t, res8)
	assert.Equal(t, "delicious", d8.Get("apples").MustString())
	assert.Equal(t, "disguisting", d8.Get("pears").MustString())

	// should override values
	res9, _ := pSocket.SimRequest("playerSettingsSet", `{"key": "apples", "value": "rotten"}`)
	testhelpers.UnpackSuccessResponse(t, res9)

	res10, _ := pSocket.SimRequest("playerSettingsGet", `{"key": "apples"}`)
	d10 := testhelpers.UnpackSuccessResponse(t, res10)
	assert.Equal(t, "rotten", d10.Get("apples").MustString())
}

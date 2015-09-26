// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"github.com/TF2Stadium/Helen/testhelpers"
	"testing"
	//	"github.com/TF2Stadium/Helen/helpers"
	//	"github.com/stretchr/testify/assert"
	"fmt"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/stretchr/testify/assert"
)

func TestChangeRole(t *testing.T) {
	testhelpers.CleanupDB()

	player1 := testhelpers.CreatePlayer()
	pSocket1 := testhelpers.NewFakeSocket()
	pSocket1.FakeAuthenticate(player1)
	SocketInit(pSocket1)

	assert.Equal(t, helpers.RolePlayer, player1.Role)

	// players can't change roles
	res1, _ := pSocket1.SimRequest("adminChangeRole", fmt.Sprintf(`{"steamid": %s, "role": "%d"}`,
		player1.SteamId, helpers.RoleMod))
	testhelpers.UnpackFailureResponse(t, res1)
	resPlayer, _ := models.GetPlayerBySteamId(player1.SteamId)
	assert.Equal(t, helpers.RolePlayer, resPlayer.Role)

	player2 := testhelpers.CreatePlayerAdmin()
	pSocket2 := testhelpers.NewFakeSocket()
	pSocket2.FakeAuthenticate(player2)
	SocketInit(pSocket2)

	// admins can change roles
	res2, _ := pSocket2.SimRequest("adminChangeRole", fmt.Sprintf(`{"steamid": "%s", "role": "moderator"}`,
		player1.SteamId, helpers.RoleMod))
	testhelpers.UnpackSuccessResponse(t, res2)
	resPlayer, _ = models.GetPlayerBySteamId(player1.SteamId)
	assert.Equal(t, helpers.RoleMod, resPlayer.Role)

	// even admins can't change roles to admin
	res3, _ := pSocket2.SimRequest("adminChangeRole", fmt.Sprintf(`{"steamid": "%s", "role": "admin"}`,
		player1.SteamId, helpers.RoleMod))
	testhelpers.UnpackFailureResponse(t, res3)
	resPlayer, _ = models.GetPlayerBySteamId(player1.SteamId)
	assert.Equal(t, helpers.RoleMod, resPlayer.Role)
}

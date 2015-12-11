// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import "github.com/TF2Stadium/Helen/helpers/authority"

// DO NOT CHANGE THE INTEGER VALUES OF ALREADY EXISTING ROLES.
const (
	RolePlayer authority.AuthRole = iota
	RoleMod
	RoleAdmin
)

var RoleNames = map[authority.AuthRole]string{
	RolePlayer: "player",
	RoleMod:    "moderator",
	RoleAdmin:  "administrator",
}

var RoleMap = map[string]authority.AuthRole{
	"player":        RolePlayer,
	"moderator":     RoleMod,
	"administrator": RoleAdmin,
}

// You cant's change the order of these
const (
	ActionBanJoin authority.AuthAction = iota
	ActionBanCreate
	ActionBanChat

	ActionChangeRole
)

var ActionNames = map[authority.AuthAction]string{
	ActionBanCreate: "ActionBanCreate",
	ActionBanJoin:   "ActionBanJoin",
	ActionBanChat:   "ActionBanChat",

	ActionChangeRole: "ActionChangeRole",
}

func InitAuthorization() {
	RoleMod.Inherit(RolePlayer)
	RoleMod.Allow(ActionBanChat)
	RoleMod.Allow(ActionBanJoin)
	RoleMod.Allow(ActionBanCreate)

	RoleAdmin.Inherit(RoleMod)
	RoleAdmin.Allow(ActionChangeRole)
}

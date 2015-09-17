// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import "github.com/TF2Stadium/Helen/helpers/authority"

// DO NOT CHANGE THE INTEGER VALUES OF ALREADY EXISTING ROLES.
const (
	RolePlayer authority.AuthRole = 0
	RoleMod    authority.AuthRole = 1
	RoleAdmin  authority.AuthRole = 2
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
	ActionBanPlayer  authority.AuthAction = iota
	ActionChangeRole authority.AuthAction = iota
)

var ActionNames = map[authority.AuthAction]string{
	ActionBanPlayer:  "ActionBanPlayer",
	ActionChangeRole: "ActionChangeRole",
}

func RoleExists(role authority.AuthRole) bool {
	_, ok := RoleNames[role]
	return ok
}

func ActionExists(role authority.AuthAction) bool {
	_, ok := ActionNames[role]
	return ok
}

func InitAuthorization() {
	RoleMod.Inherit(RolePlayer)
	RoleMod.Allow(ActionBanPlayer)

	RoleAdmin.Inherit(RoleMod)
	RoleAdmin.Allow(ActionChangeRole)
}

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
	RoleDeveloper
)

var RoleNames = map[authority.AuthRole]string{
	RoleDeveloper: "developer",
	RolePlayer:    "player",
	RoleMod:       "moderator",
	RoleAdmin:     "administrator",
}

var RoleMap = map[string]authority.AuthRole{
	"player":        RolePlayer,
	"moderator":     RoleMod,
	"administrator": RoleAdmin,
	"developer":     RoleDeveloper,
}

// You can't change the order of these
const (
	ActionBanJoin authority.AuthAction = iota
	ActionBanCreate
	ActionBanChat
	ActionChangeRole
	ActionViewLogs
	ActionViewPage //view admin pages
	ActionDeleteChat
	ModifyServers //add/remove servers
)

var ActionNames = map[authority.AuthAction]string{
	ActionBanCreate: "ActionBanCreate",
	ActionBanJoin:   "ActionBanJoin",
	ActionBanChat:   "ActionBanChat",

	ActionChangeRole: "ActionChangeRole",
}

func InitAuthorization() {
	RoleDeveloper.Allow(ActionViewPage)

	RoleMod.Inherit(RolePlayer)
	RoleMod.Allow(ActionBanChat)
	RoleMod.Allow(ActionBanJoin)
	RoleMod.Allow(ActionBanCreate)
	RoleMod.Allow(ActionViewLogs)
	RoleMod.Allow(ActionViewPage)
	RoleMod.Allow(ActionDeleteChat)
	RoleMod.Allow(ModifyServers)

	RoleAdmin.Inherit(RoleMod)
	RoleAdmin.Allow(ActionChangeRole)
}

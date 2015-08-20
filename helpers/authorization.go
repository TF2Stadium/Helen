package helpers

import "github.com/TF2Stadium/Helen/helpers/authority"

// DO NOT CHANGE THE NUMBERS OF ALREADY EXISTING ROLES.
const (
	RolePlayer authority.AuthRole = 0
	RoleMod    authority.AuthRole = 1
	RoleAdmin  authority.AuthRole = 2
)

var RoleNames = map[authority.AuthRole]string{
	0: "Player",
	1: "Moderator",
	2: "Administrator",
}

const (
	ActionBanPlayer authority.AuthAction = 1
)

var ActionNames = map[authority.AuthAction]string{
	ActionBanPlayer: "ActionBanPlayer",
}

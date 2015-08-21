package helpers

import "github.com/TF2Stadium/Helen/helpers/authority"

// DO NOT CHANGE THE INTEGER VALUES OF ALREADY EXISTING ROLES.
const (
	RolePlayer authority.AuthRole = 0
	RoleMod    authority.AuthRole = 1
	RoleAdmin  authority.AuthRole = 2
)

var RoleNames = map[authority.AuthRole]string{
	RolePlayer: "Player",
	RoleMod:    "Moderator",
	RoleAdmin:  "Administrator",
}

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

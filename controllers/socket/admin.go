package socket

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
)

var changeRoleParams = map[string]chelpers.Param{
	"steamid": chelpers.Param{Type: chelpers.PTypeString},
	"role":    chelpers.Param{Type: chelpers.PTypeInt},
}

func ChangeRole(socketid string) func(string) string {
	return chelpers.AuthorizationFilter(socketid, helpers.ActionChangeRole,
		chelpers.JsonVerifiedFilter(changeRoleParams,
			func(js *simplejson.Json) string {
				roleInt, _ := js.Get("role").Int()
				role := authority.AuthRole(roleInt)
				if !helpers.RoleExists(role) || role == helpers.RoleAdmin {
					bytes, _ := chelpers.BuildFailureJSON("Invalid role parameter.", 0).Encode()
					return string(bytes)
				}

				steamid, _ := js.Get("steamid").String()
				otherPlayer, err := models.GetPlayerBySteamId(steamid)
				if err != nil {
					bytes, _ := chelpers.BuildFailureJSON("Player not found.", 0).Encode()
					return string(bytes)
				}

				currPlayer, _ := chelpers.GetPlayerSocket(socketid)

				models.LogAdminAction(currPlayer.ID, helpers.ActionChangeRole, otherPlayer.ID)

				otherPlayer.Role = role
				db.DB.Save(&otherPlayer)
				return chelpers.BuildEmptySuccessString()
			}))
}

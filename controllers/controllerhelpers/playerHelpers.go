package controllerhelpers

import (
	"time"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
)

var banTypeList = []string{"play", "create", "chat", "full"}

func getPlayerBanPointer(name string, player *models.Player) *int64 {
	switch name {
	case "play":
		return &player.BannedPlayUntil
	case "create":
		return &player.BannedCreateUntil
	case "chat":
		return &player.BannedChatUntil
	case "full":
		return &player.BannedFullUntil
	default:
		return nil
	}
}

func GetPlayerBanTimes(player *models.Player) map[string]int64 {
	res := make(map[string]int64)
	for _, typ := range banTypeList {
		res[typ] = *getPlayerBanPointer(typ, player)
	}

	return res
}

func UnbanPlayer(name string, player *models.Player) *helpers.TPError {
	return SetPlayerBanTime(0, name, player)
}

func SetPlayerBanTime(i int64, name string, player *models.Player) *helpers.TPError {
	pnt := getPlayerBanPointer(name, player)
	if pnt == nil {
		return helpers.NewTPError("Invalid ban type", -1)
	}

	*pnt = i
	return helpers.NewTPErrorFromError(player.Save())
}

func IsPlayerBanActive(i int64) bool {
	return i > time.Now().Unix()
}

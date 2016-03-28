package handler

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/wsevent"
)

//Mumble object contains methods for changing mumble details for a user
type Mumble struct{}

func (Mumble) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Mumble) ResetMumblePassword(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)
	player.MumbleAuthkey = player.GenAuthKey()
	player.Save()

	return emptySuccess
}

func (Mumble) GetMumblePassword(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)

	return newResponse(struct {
		Password string `json:"password"`
	}{player.MumbleAuthkey})
}

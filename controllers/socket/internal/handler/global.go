// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
	"github.com/bitly/go-simplejson"
)

type Global struct{}

func (Global) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Global) GetConstant(_ *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	var args struct {
		Constant string `json:"constant"`
	}
	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err)
	}

	output := simplejson.New()
	switch args.Constant {
	case "lobbySettingsList":
		output = models.LobbySettingsToJson()
	default:
		return helpers.NewTPError("Unknown constant.", -1)
	}

	return chelpers.BuildSuccessJSON(output)
}

func (Global) GetSocketInfo(server *wsevent.Server, so *wsevent.Client, data []byte) interface{} {
	socketinfo := struct {
		ID    string   `json:"id"`
		Rooms []string `json:"rooms"`
	}{so.Id(), server.RoomsJoined(so.Id())}

	return chelpers.BuildSuccessJSON(socketinfo)
}

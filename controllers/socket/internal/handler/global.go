// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package handler

import (
	"encoding/json"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
	"github.com/bitly/go-simplejson"
)

func GetConstant(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	var args struct {
		Constant string `json:"constant"`
	}
	if err := chelpers.GetParams(data, &args); err != nil {
		return helpers.NewTPErrorFromError(err).Encode()
	}

	output := simplejson.New()
	switch args.Constant {
	case "lobbySettingsList":
		output = models.LobbySettingsToJson()
	default:
		return helpers.NewTPError("Unknown constant.", -1).Encode()
	}

	bytes, _ := chelpers.BuildSuccessJSON(output).Encode()
	return bytes
}

func GetSocketInfo(server *wsevent.Server, so *wsevent.Client, data []byte) []byte {
	socketinfo := struct {
		ID    string   `json:"id"`
		Rooms []string `json:"rooms"`
	}{so.Id(), server.RoomsJoined(so.Id())}

	bytes, _ := json.Marshal(socketinfo)
	return bytes
}

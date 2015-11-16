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

func GetConstant(server *wsevent.Server, so *wsevent.Client, data string) string {
	var args struct {
		Constant string `json:"constant"`
	}
	if err := chelpers.GetParams(data, &args); err != nil {
		bytes, _ := helpers.NewTPErrorFromError(err).Encode()
		return string(bytes)
	}

	output := simplejson.New()
	switch args.Constant {
	case "lobbySettingsList":
		output = models.LobbySettingsToJson()
	default:
		bytes, _ := helpers.NewTPError("Unknown constant.", -1).Encode()
		return string(bytes)
	}

	bytes, _ := chelpers.BuildSuccessJSON(output).Encode()
	return string(bytes)
}

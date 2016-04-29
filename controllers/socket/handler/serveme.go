package handler

import (
	"errors"
	"strings"

	"github.com/Sirupsen/logrus"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/gameserver"
	"github.com/TF2Stadium/servemetf"
	"github.com/TF2Stadium/wsevent"
)

type Serveme struct{}

func (Serveme) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Serveme) GetServemeServers(so *wsevent.Client, _ struct{}) interface{} {
	context := helpers.GetServemeContextIP(chelpers.GetIPAddr(so.Request))

	starts, ends, err := context.GetReservationTime(so.Token.Claims["steam_id"].(string))
	if err != nil {
		logrus.Error(err)
		return errors.New("Cannot access serveme.tf")
	}

	reservations, err := context.FindServers(starts, ends, so.Token.Claims["steam_id"].(string))
	if err != nil {
		logrus.Error(err)
		return errors.New("Cannot access serveme.tf")
	}

	var servers []servemetf.Server

	for _, server := range reservations.Servers {
		//Out of respect for TF2Center, we don't use their servers with serveme integration.
		if strings.HasPrefix(server.Name, "TF2Center") {
			continue
		}
		servers = append(servers, server)
	}

	resp := struct {
		StartsAt string             `json:"startsAt"`
		EndsAt   string             `json:"endsAt"`
		Servers  []servemetf.Server `json:"servers"`
	}{starts.Format(servemetf.TimeFormat), ends.Format(servemetf.TimeFormat), servers}

	return newResponse(resp)
}

func (Serveme) GetStoredServers(so *wsevent.Client, _ struct{}) interface{} {
	servers := gameserver.GetAvailableServers()
	return newResponse(servers)
}

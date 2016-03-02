package handler

import (
	"errors"
	"net"
	"strings"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/servemetf"
	"github.com/TF2Stadium/wsevent"
)

type Serveme struct{}

func (Serveme) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Serveme) GetServemeServers(so *wsevent.Client, _ struct{}) interface{} {
	starts, ends, err := helpers.ServemeContext.GetReservationTime(so.Token.Claims["steam_id"].(string))
	if _, ok := err.(*net.OpError); ok {
		return errors.New("Cannot access serveme.tf")
	}

	reservations, err := helpers.ServemeContext.FindServers(starts, ends, so.Token.Claims["steam_id"].(string))
	if _, ok := err.(*net.OpError); ok {
		return errors.New("Cannot access serveme.tf")
	}

	for i, server := range reservations.Servers {
		//Out of respect for TF2Center, we don't use their servers with serveme integration.
		if strings.HasPrefix(server.Name, "TF2Center") {
			reservations.Servers[i] = reservations.Servers[len(reservations.Servers)-1]
			reservations.Servers = reservations.Servers[:len(reservations.Servers)-1]
		}
	}

	resp := struct {
		StartsAt string             `json:"startsAt"`
		EndsAt   string             `json:"endsAt"`
		Servers  []servemetf.Server `json:"servers"`
	}{starts.Format(servemetf.TimeFormat), ends.Format(servemetf.TimeFormat), reservations.Servers}

	return newResponse(resp)
}

func (Serveme) GetStoredServers(so *wsevent.Client, _ struct{}) interface{} {
	servers := models.GetAvailableServers()
	return newResponse(servers)
}

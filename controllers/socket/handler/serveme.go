package handler

import (
	"errors"
	"net"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/servemetf"
	"github.com/TF2Stadium/wsevent"
)

type Serveme struct{}

func (Serveme) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Serveme) GetServemeServers(so *wsevent.Client, _ struct{}) interface{} {
	starts, ends, err := helpers.ServemeContext.GetReservationTime()
	if _, ok := err.(*net.OpError); ok {
		return errors.New("Cannot access serveme.tf")
	}

	reservations, err := helpers.ServemeContext.FindServers(starts, ends)
	if _, ok := err.(*net.OpError); ok {
		return errors.New("Cannot access serveme.tf")
	}

	resp := struct {
		StartsAt string             `json:"startsAt"`
		EndsAt   string             `json:"endsAt"`
		Servers  []servemetf.Server `json:"servers"`
	}{starts.Format(servemetf.TimeFormat), ends.Format(servemetf.TimeFormat), reservations.Servers}

	return newResponse(resp)
}

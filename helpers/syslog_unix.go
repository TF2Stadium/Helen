//+build !windows

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.
package helpers

import (
	"log/syslog"

	"github.com/op/go-logging"
)

func setupPapertrail(addr string, backend logging.Backend) {
	writer, err := syslog.Dial("udp4", addr, syslog.LOG_EMERG, "Helen")
	if err != nil {
		Logger.Fatal(err.Error())
	}

	format = logging.MustStringFormatter(`[%{level:.4s}] %{shortfile} %{shortfunc}() : %{message}`)
	syslogBackend := logging.NewBackendFormatter(&logging.SyslogBackend{Writer: writer}, format)
	logging.SetBackend(backend, syslogBackend)
}

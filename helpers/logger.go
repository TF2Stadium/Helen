// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"fmt"
	"log/syslog"
	"os"

	"github.com/op/go-logging"
)

type FakeLogger struct{}

func (f FakeLogger) Print(v ...interface{}) {
	Logger.Warning(fmt.Sprint(v))
}

var Logger = logging.MustGetLogger("main")

var format = logging.MustStringFormatter(
	`%{time:15:04:05} %{color} [%{level:.4s}] %{shortfunc}() : %{message} %{color:reset}`)

var syslogFormat = logging.MustStringFormatter(
	`%{color} [%{level:.4s}] %{shortfunc}() : %{message} %{color:reset}`)

// Sample usage
// Logger.Debug("debug %s", Password("secret"))
// Logger.Info("info")
// Logger.Notice("notice")
// Logger.Warning("warning")
// Logger.Error("err")
// Logger.Critical("crit")

var syslogAddr = os.Getenv("SYSLOG_ADDR")

func InitLogger() {
	console := logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), format)

	if syslogAddr != "" {
		writer, err := syslog.Dial("udp", syslogAddr, syslog.LOG_ALERT, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		backend := &logging.SyslogBackend{writer}

		backendFormatter := logging.NewBackendFormatter(backend, syslogFormat)
		logging.SetBackend(backendFormatter, console)
	} else {
		logging.SetBackend(console)
	}
}

// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
)

var Logger = logging.MustGetLogger("main")

var format = logging.MustStringFormatter(
	`%{time:15:04:05} %{color} [%{level:.4s}] %{shortfile} %{shortfunc}() : %{message} %{color:reset}`)

type FakeLogger struct{}

func (f FakeLogger) Print(v ...interface{}) {
	Logger.Warning(fmt.Sprint(v))
}

func InitLogger() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if addr := os.Getenv("PAPERTRAIL_ADDR"); addr != "" {
		setupPapertrail(addr, backendFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}
}

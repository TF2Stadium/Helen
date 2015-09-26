// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"os"

	"fmt"
	"github.com/op/go-logging"
)

type FakeLogger struct{}

func (f FakeLogger) Print(v ...interface{}) {
	Logger.Warning(fmt.Sprint(v))
}

var Logger = logging.MustGetLogger("main")

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

// Sample usage
// Logger.Debug("debug %s", Password("secret"))
// Logger.Info("info")
// Logger.Notice("notice")
// Logger.Warning("warning")
// Logger.Error("err")
// Logger.Critical("crit")

func InitLogger() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter)
}

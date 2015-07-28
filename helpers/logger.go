package helpers

import (
	"os"

	"github.com/op/go-logging"
)

var Logger = logging.MustGetLogger("example")

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

// Sample usage
// log.Debug("debug %s", Password("secret"))
// log.Info("info")
// log.Notice("notice")
// log.Warning("warning")
// log.Error("err")
// log.Critical("crit")

func InitLogger() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)
}

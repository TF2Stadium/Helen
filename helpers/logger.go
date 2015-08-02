package helpers

import (
	"os"

	"github.com/op/go-logging"
)

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

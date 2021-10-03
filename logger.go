package main

import (
	"os"

	"github.com/op/go-logging"
)

var Logger = logging.MustGetLogger("app")

func SetupLogger() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	format := logging.MustStringFormatter(
		"%{time:15:04:05.000}\t%{level:.4s}\t%{shortfunc}: %{message}",
	)
	formatter := logging.NewBackendFormatter(backend, format)

	backendLeveled := logging.AddModuleLevel(formatter)
	backendLeveled.SetLevel(logging.DEBUG, "")

	logging.SetBackend(backendLeveled)

	Logger.Debug("Logger configured")
}

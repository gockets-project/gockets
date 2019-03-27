package ll

import (
	"github.com/op/go-logging"
	"os"
)

var Log *logging.Logger

var backendLevels = map[int]logging.Level{
	1: logging.INFO,
	2: logging.CRITICAL,
	3: logging.DEBUG,
}

func InitLogger(logLevel int) {
	Log = logging.MustGetLogger("main")
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	backend := logging.NewBackendFormatter(logBackend, format)
	leveledBackend := logging.AddModuleLevel(backend)
	leveledBackend.SetLevel(backendLevels[logLevel], "main")

	logging.SetBackend(leveledBackend)
	Log.Debugf("Debug level is: %s", leveledBackend.GetLevel("main"))
	Log.Debug("Logger started")
}

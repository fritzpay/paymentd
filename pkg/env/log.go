package env

import (
	"gopkg.in/inconshreveable/log15.v2"
	golog "log"
	"os"
)

var Log log15.Logger

// logBridge acts as a Writer for the log pkg
// It will log to log15
type logBridge struct {
	log log15.Logger
}

// logBridge Writer implementation
// will log all log pkg messages as log15.Info messages
func (l logBridge) Write(msg []byte) (int, error) {
	l.log.Info("log pkg message", log15.Ctx{"message": string(msg)})
	return len(msg), nil
}

// InitLog adjusts the logging environment and returns a log15.Logger for further
// use.
func InitLog() log15.Logger {
	Log = log15.New()
	Log.SetHandler(log15.StreamHandler(os.Stdout, log15.LogfmtFormat()))
	golog.SetOutput(logBridge{Log})
	return Log
}

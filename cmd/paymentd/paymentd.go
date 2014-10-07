package main

import (
	"flag"
	"gopkg.in/inconshreveable/log15.v2"
	golog "log"
	"os"
)

const (
	// AppName is the name of the application
	AppName = "paymentd"
	// AppVersion is the version of the application
	AppVersion = "0.1"
)

// command line flags
var (
	// cfgFileName is the configuration file to use
	cfgFileName string
)

var (
	log log15.Logger
)

func main() {
	// set flags
	flag.StringVar(&cfgFileName, "c", "", "config file name to use")
}

// logBridge acts as a Writer for the log pkg
// It will log to log15
type logBridge int

// logBridge Writer implementation
// will log all log pkg messages as log15.Info messages
func (l logBridge) Write(msg []byte) (int, error) {
	log.Info("log pkg message", log15.Ctx{"message": string(msg)})
	return len(msg), nil
}

// setLog sets up the general logging to stdout using log15 <gopkg.in/inconshreveable/log15.v2>
// it also re-routes the log pkg output to log15 in case other packages log directly through
// the log pkg
func setLog() {
	log = log15.New(log15.Ctx{
		"AppName":    AppName,
		"AppVersion": AppVersion,
		"PID":        os.Getpid(),
	})
	log.SetHandler(log15.StreamHandler(os.Stdout, log15.LogfmtFormat()))
	golog.SetOutput(logBridge(0))
}

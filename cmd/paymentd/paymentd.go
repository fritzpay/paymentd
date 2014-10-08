package main

import (
	"flag"
	"github.com/fritzpay/paymentd/pkg/env"
	"gopkg.in/inconshreveable/log15.v2"
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

	log = env.Log.New(log15.Ctx{
		"AppName":    AppName,
		"AppVersion": AppVersion,
		"PID":        os.Getpid(),
	})
}

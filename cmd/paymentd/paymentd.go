package main

import (
	"flag"
	"github.com/fritzpay/paymentd/pkg/config"
	"github.com/fritzpay/paymentd/pkg/env"
	"github.com/fritzpay/paymentd/pkg/server"
	"github.com/fritzpay/paymentd/pkg/service/api"
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
	cfg config.Config
	srv *server.Server
)

func main() {
	// set flags
	flag.StringVar(&cfgFileName, "c", "", "config file name to use")
	flag.Parse()

	log = env.Log.New(log15.Ctx{
		"AppName":    AppName,
		"AppVersion": AppVersion,
		"PID":        os.Getpid(),
	})
	log.Info("starting daemon...")

	log.Info("loading config...")
	loadConfig()

	srv = server.NewServer()

	// API handler
	if cfg.API.Active {
		log.Info("enabling API service...")
		apiHandler, err := api.NewHandler(cfg)
		if err != nil {
			log.Crit("error initializing API service", log15.Ctx{"err": err})
			log.Info("exiting...")
			os.Exit(1)
		}
		srv.RegisterService(cfg.API.Service, apiHandler)
	}

	err := srv.Serve()
	if err != nil {
		log.Crit("error serving", log15.Ctx{"err": err})
		log.Info("exiting...")
		os.Exit(1)
	}
}

func loadConfig() {
	if cfgFileName == "" {
		log.Info("no config file provided. trying default config...")
		cfg = config.DefaultConfig()
	} else {
		cfgFile, err := os.Open(cfgFileName)
		if err != nil {
			log.Crit("could not open config file", log15.Ctx{"cfgFileName": cfgFileName})
			log.Info("exiting...")
			os.Exit(1)
		}
		cfg, err = config.ReadConfig(cfgFile)
		if err != nil {
			log.Crit("error reading config file", log15.Ctx{"cfgFileName": cfgFileName, "err": err})
			log.Info("exiting...")
			os.Exit(1)
		}
		err = cfgFile.Close()
		if err != nil {
			log.Crit("error closing config file", log15.Ctx{"cfgFileName": cfgFileName, "err": err})
			log.Info("exiting...")
			os.Exit(1)
		}
	}
}

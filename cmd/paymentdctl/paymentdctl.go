package main

import (
	"os"

	"github.com/codegangsta/cli"
)

const (
	// AppName is the name of the application
	AppName = "paymentdctl"
	// AppVersion is the version of the application
	AppVersion = "0.1"
	// AppDescription describes what this application does
	AppDescription = "paymentd c&c and utilities"
)

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Version = AppVersion
	app.Usage = AppDescription

	app.Commands = []cli.Command{
		configCommand,
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "config file name",
		},
	}

	app.Run(os.Args)
}

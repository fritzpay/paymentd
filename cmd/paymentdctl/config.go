package main

import (
	"fmt"
	"net"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fritzpay/paymentd/pkg/config"
)

const cfgCommandDescription = `This command allows you to load, modify and test configuration 
files.`

var cfg = config.DefaultConfig()

var configCommand = cli.Command{
	Name:        "config",
	ShortName:   "cfg",
	Usage:       "Configuration related tools.",
	Description: cfgCommandDescription,
	Subcommands: []cli.Command{
		testConfigComand,
		writeConfigCommand,
	},
}

var testConfigComand = cli.Command{
	Name:      "test",
	ShortName: "t",
	Usage:     "Test configuration.",
	Action:    testConfigAction,
}

func configFileName(c *cli.Context) string {
	return c.GlobalString("config")
}

func readConfig(c *cli.Context) bool {
	if configFileName(c) != "" {
		if !readConfigFile(configFileName(c)) {
			return false
		}
	} else {
		fmt.Println("no config file flag provided. will use default config...")
	}
	return true
}

func readConfigFile(cfgFileName string) bool {
	fmt.Printf("will read config file %s...\n", cfgFileName)
	cfgFile, err := os.Open(cfgFileName)
	if err != nil {
		fmt.Printf("error opening config file %s: %v\n", cfgFileName, err)
		return false
	}
	defer cfgFile.Close()
	cfg, err = config.ReadConfig(cfgFile)
	if err != nil {
		fmt.Printf("error reading config file %s: %v\n", cfgFileName, err)
		return false
	}
	return true
}

func testConfigAction(c *cli.Context) {
	if !readConfig(c) {
		return
	}

	errors := 0
	warnings := 0

	if cfg.API.Service.Address == "" {
		errors++
		fmt.Println("error: API.Service.Address is empty.")
	} else {
		apiAddr, err := net.ResolveTCPAddr("tcp", cfg.API.Service.Address)
		if err != nil {
			errors++
			fmt.Printf("error: API.Service.Address could not be resolved: %v\n", err)
		}
		fmt.Printf("API.Service.Address: API server will use network %s and address %s\n", apiAddr.Network(), apiAddr.String())
	}

	fmt.Printf("\n\nconfig testing complete.\n%d errors and %d warnings.\n", errors, warnings)
}

var writeConfigCommand = cli.Command{
	Name:      "write",
	ShortName: "w",
	Usage:     "Will write the config in buffer to the given output file.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "output, o",
			Usage: "Output file to write to.",
		},
	},
	Action: writeConfigAction,
}

func writeConfigAction(c *cli.Context) {
	cfgFileName := c.String("output")
	if cfgFileName == "" {
		fmt.Print("no output file name provided\n\n")
		cli.ShowCommandHelp(c, "w")
		return
	}

	if !readConfig(c) {
		return
	}
	cfgFile, err := os.OpenFile(cfgFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("error opening config file %s for writing: %v\n", cfgFileName, err)
		return
	}
	defer cfgFile.Close()
	err = config.WriteConfig(cfgFile, cfg)
	if err != nil {
		fmt.Printf("error writing config file %s: %v\n", cfgFileName, err)
		return
	}
	fmt.Printf("config file %s written.\n", cfgFileName)
}

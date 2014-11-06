// +build ignore

/*
Copyright 2014 The FritzPay Authors.
*/

/*
This program builds paymentd
*/
package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
)

var (
	verbose bool
	quiet   bool
	test    bool
	rebuild bool
)

var output bytes.Buffer

var dependencies = []string{
	"golang.org/x/tools/cmd/vet",
	"github.com/tools/godep",
}

var installPkgs = []string{
	"github.com/fritzpay/paymentd/cmd/paymentd",
	"github.com/fritzpay/paymentd/cmd/paymentdctl",
}

func main() {
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.BoolVar(&quiet, "q", false, "do not print anything unless there is an error")
	flag.BoolVar(&test, "t", true, "execute tests before building")
	flag.BoolVar(&rebuild, "r", false, "force rebuilding binaries")
	flag.Parse()

	log.SetFlags(0)

	installDependencies()
	if test {
		runTests()
	}
	installBinaries()

	if !quiet {
		log.Print("Success.")
	}
}

func setCmdIO(cmd *exec.Cmd) {
	if quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
}

func installDependencies() {
	args := []string{
		"get",
	}
	if verbose {
		args = append(args, "-v")
		log.Print("installing dependencies...")
	}
	for _, pkg := range dependencies {
		a := append(args, pkg)
		cmd := exec.Command("go", a...)
		setCmdIO(cmd)
		if err := cmd.Run(); err != nil {
			log.Printf("error installing dependencies: %v\n%s", err, output.String())
			os.Exit(1)
		}
	}
	args = []string{
		"restore",
		"./...",
	}
	if verbose {
		log.Print("resolving dependencies...")
	}
	cmd := exec.Command("godep", args...)
	setCmdIO(cmd)
	if err := cmd.Run(); err != nil {
		log.Printf("error resolving dependencies: %v\n%s", err, output.String())
		os.Exit(1)
	}
}

func runTests() {
	cmds := [][]string{
		[]string{
			"vet",
		},
		[]string{
			"test",
		},
	}
	if verbose {
		log.Print("running tests...")
		cmds[0] = append(cmds[0], "-x")
		cmds[1] = append(cmds[1], "-v")
	}
	for _, cArgs := range cmds {
		a := append(cArgs, "./...")
		cmd := exec.Command("go", a...)
		setCmdIO(cmd)
		if err := cmd.Run(); err != nil {
			log.Printf("error on tests: %v\n%s", err, output.String())
			os.Exit(2)
		}
	}
}

func installBinaries() {
	args := []string{
		"install",
	}
	if verbose {
		args = append(args, "-v", "-x")
	}
	if rebuild {
		args = append(args, "-a")
	}
	for _, install := range installPkgs {
		a := append(args, install)
		cmd := exec.Command("go", a...)
		setCmdIO(cmd)
		if err := cmd.Run(); err != nil {
			log.Printf("error building binary %s: %v\n%s", install, err, output.String())
			os.Exit(1)
		}
	}
}

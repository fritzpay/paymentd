// +build ignore

/*
Copyright 2014 The FritzPay Authors.
*/

/*
This program builds paymentd
*/
package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetFlags(0)
	args := []string{
		"get",
		"-v",
	}
	pkgs := []string{
		"code.google.com/p/go.tools/cmd/vet",
		"github.com/tools/godep",
	}
	for _, pkg := range pkgs {
		a := append(args, pkg)
		cmd := exec.Command("go", a...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}
	args = []string{
		"vet",
		"./...",
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

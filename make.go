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
	"path/filepath"
)

const (
	paymentdPkg = "github.com/fritzpay/paymentd"
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

var pkgPath string

var installPkgs = []string{
	paymentdPkg + "/cmd/paymentd",
	paymentdPkg + "/cmd/paymentdctl",
}

func main() {
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.BoolVar(&quiet, "q", false, "do not print anything unless there is an error")
	flag.BoolVar(&test, "t", true, "execute tests before building")
	flag.BoolVar(&rebuild, "r", false, "force rebuilding binaries")
	flag.Parse()

	log.SetFlags(0)

	var err error
	pkgPath, err = goPkgPath(paymentdPkg)
	if err != nil {
		log.Printf("error determining pkg path: %v", err)
		os.Exit(1)
	}

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
		"get",
		"./...",
	}
	if verbose {
		log.Print("resolving dependencies...")
	}
	godepBin, err := goBin("godep")
	if err != nil {
		log.Printf("error on Godep install. cannot find binary: %v", err)
		os.Exit(1)
	}
	cmd := exec.Command(godepBin, args...)
	cmd.Dir = pkgPath
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
		cmd.Dir = pkgPath
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

func goBin(bin string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", os.ErrNotExist
	}
	for _, p := range filepath.SplitList(gopath) {
		dir := filepath.Join(p, "bin", filepath.FromSlash(bin))
		fi, err := os.Stat(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if m := fi.Mode(); !m.IsDir() && m&0111 != 0 {
			return dir, nil
		}
	}
	return "", os.ErrNotExist
}

func goPkgPath(pkg string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", os.ErrNotExist
	}
	for _, p := range filepath.SplitList(gopath) {
		dir := filepath.Join(p, "src", filepath.FromSlash(pkg))
		fi, err := os.Stat(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !fi.IsDir() {
			continue
		}
		return dir, nil
	}
	return "", os.ErrNotExist
}

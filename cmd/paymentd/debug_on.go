// +build debug

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/fritzpay/paymentd/pkg/env"
)

func setEnv() {
	env.SetRuntime()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	go func() {
		<-sigs
		panic("trace")
	}()
}

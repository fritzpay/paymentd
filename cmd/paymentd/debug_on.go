// +build debug

package main

import (
	"github.com/fritzpay/paymentd/pkg/env"
	"os"
	"os/signal"
	"syscall"
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

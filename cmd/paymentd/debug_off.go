// +build !debug

package main

import (
	"github.com/fritzpay/paymentd/pkg/env"
)

func setEnv() {
	env.SetRuntime()
}

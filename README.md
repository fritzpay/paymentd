[![Build Status](https://travis-ci.org/fritzpay/paymentd.svg?branch=master)](https://travis-ci.org/fritzpay/paymentd)

paymentd
========

FritzPay paymentd is the payment server of the FritzPay stack.

# Install

Retrieve the sources for paymentd.

`$ go get -d github.com/fritzpay/paymentd`

You can use the `make.go` script.

```
$ cd $GOPATH/src/github.com/fritzpay/paymentd
$ go run make.go
```

## Manual Install:

If you don't have [Godep](https://github.com/tools/godep) installed.

If you have Godep installed, you can skip this step.

`$ go install github.com/tools/godep`

Restore the dependencies.

```
$ cd $GOPATH/src/github.com/fritzpay/paymentd
$ godep restore ./...
```

And build paymentd.

`$ go install github.com/fritzpay/paymentd/cmd/paymentd`

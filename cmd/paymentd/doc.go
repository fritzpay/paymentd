/*
Copyright 2014 The FritzPay Authors.
*/

/*
The paymentd daemon serves payment related services for the FritzPay stack.

Usage:
  paymentd

  Flags understood by paymentd:
    -c          Path to config file name.
                Alternatively the environment var $PAYMENTDCFG can be used to set
                the configuration file name.

  Example:
    paymentd -c /etc/paymentd/paymentd.config.json
*/
package main

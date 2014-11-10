// +build go1.3

package fritzpay

import (
	"net/http"
	"time"
)

func newClient() (*http.Transport, *http.Client) {
	tr := &http.Transport{}
	cl := &http.Client{
		Transport: tr,
		Timeout:   time.Second,
	}
	return tr, cl
}

// +build !go1.3

package fritzpay

import (
	"net/http"
)

func newClient() (*http.Transport, *http.Client) {
	tr := &http.Transport{}
	cl := &http.Client{
		Transport: tr,
	}
	return tr, cl
}

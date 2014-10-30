package v1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"hash"
	"net/http"
	"strconv"
	"time"
)

// GetPaymentRequest represents a get payment request
type GetPaymentRequest struct {
	ProjectKey   string
	PaymentId    string
	paymentID    payment.PaymentID
	Ident        string
	Timestamp    int64
	Nonce        string
	hexSignature string
}

func (r *GetPaymentRequest) Message() ([]byte, error) {
	var err error
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(r.ProjectKey)
	if err != nil {
		return nil, fmt.Errorf("buffer error: %v", err)
	}
	if r.PaymentId != "" {
		_, err = buf.WriteString(r.PaymentId)
		if err != nil {
			return nil, fmt.Errorf("buffer error: %v", err)
		}
	} else if r.Ident != "" {
		_, err = buf.WriteString(r.Ident)
		if err != nil {
			return nil, fmt.Errorf("buffer error: %v", err)
		}
	} else {
		return nil, fmt.Errorf("neither payment id nor ident set")
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Timestamp, 10))
	if err != nil {
		return nil, fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Nonce)
	if err != nil {
		return nil, fmt.Errorf("buffer error: %v", err)
	}
	return buf.Bytes(), nil
}

func (r *GetPaymentRequest) HashFunc() func() hash.Hash {
	return sha256.New
}

func (r *GetPaymentRequest) Signature() ([]byte, error) {
	return hex.DecodeString(r.hexSignature)
}

func (r *GetPaymentRequest) RequestProjectKey() string {
	return r.ProjectKey
}

func (r *GetPaymentRequest) Time() time.Time {
	return time.Unix(r.Timestamp, 0)
}

func (r *GetPaymentRequest) ReadFromRequest(req *http.Request) error {
	var err error
	vars := mux.Vars(req)
	if vars["paymentId"] != "" {
		r.PaymentId = vars["paymentId"]
		r.paymentID, err = payment.ParsePaymentIDStr(r.PaymentId)
		if err != nil {
			return errors.New("invalid payment id")
		}
	} else if vars["ident"] != "" {
		r.Ident = vars["ident"]
	}
	q := req.URL.Query()
	r.ProjectKey = q.Get("ProjectKey")
	if r.ProjectKey == "" {
		return errors.New("no project key")
	}
	r.Timestamp, err = strconv.ParseInt(q.Get("Timestamp"), 10, 64)
	if err != nil {
		return err
	}
	r.Nonce = q.Get("Nonce")
	if r.Nonce == "" {
		return errors.New("no nonce")
	}
	return nil
}

func (a *PaymentAPI) GetPayment() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{
			"method": "GetPayment",
		})
		var err error
		req := &GetPaymentRequest{}
		err = req.ReadFromRequest(r)
		if err != nil {
			ret := ErrReadParam
			if Debug {
				ret.Info = err.Error()
			}
			ret.Write(w)
			return
		}
		if req.PaymentId != "" {
			log = log.New(log15.Ctx{"DisplayPaymentId": req.PaymentId})
		} else if req.Ident != "" {
			log = log.New(log15.Ctx{"Ident": req.Ident})
		} else {
			ret := ErrInval
			ret.Info = "neither payment id nor ident in request"
			ret.Write(w)
			return
		}
		var projectKey *project.Projectkey
		if projectKey = a.authenticateRequest(req, log, w); projectKey == nil {
			return
		}
	})
}

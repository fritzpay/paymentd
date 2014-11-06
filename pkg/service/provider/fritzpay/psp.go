package fritzpay

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"net/url"
	"time"
)

func pspInit(ctx context.Context, fritzpayP Payment, callbackURL string) {
	if deadline, ok := ctx.Deadline(); ok {
		// let's assume we will need at least 3 seconds to run
		if deadline.Before(time.Now().Add(3 * time.Second)) {
			return
		}
	}
	log := ctx.Value("log").(log15.Logger).New(log15.Ctx{
		"pkg":         "github.com/fritzpay/paymentd/pkg/service/provider/fritzpay",
		"method":      "doInit",
		"callbackURL": callbackURL,
	})
	callback, err := url.Parse(callbackURL)
	if err != nil {
		log.Error("error on parsing callback URL", log15.Ctx{"err": err})
		return
	}
	tx, err := ctx.Value("paymentDB").(*sql.DB).Begin()
	if err != nil {
		log.Crit("error on begin tx", log15.Ctx{"err": err})
		return
	}
	if Debug {
		log.Debug("worker start...")
	}
	var req *http.Request
	tr := &http.Transport{}
	cl := &http.Client{
		Transport: tr,
		Timeout:   time.Second,
	}
	ok := make(chan struct{})
	errors := make(chan error)
	go func() {
		paymentTx, err := PaymentTransactionCurrentByPaymentIDProviderTx(tx, fritzpayP.ID)
		if err != nil && err != ErrTransactionNotFound {
			errors <- err
			return
		}
		if err == ErrTransactionNotFound {
			h := sha1.New()
			_, err = h.Write([]byte(fmt.Sprintf("%d", fritzpayP.ID)))
			if err != nil {
				errors <- err
				return
			}
			paymentTx.FritzpayPaymentID = fritzpayP.ID
			paymentTx.Timestamp = time.Now()
			paymentTx.Status = TransactionPSPInit
			paymentTx.FritzpayID.String, paymentTx.FritzpayID.Valid = hex.EncodeToString(h.Sum(nil)), true
			paymentTx.Payload.String, paymentTx.Payload.Valid = "initialized on psp", true
			err = InsertPaymentTransactionTx(tx, paymentTx)
			if err != nil {
				errors <- err
				return
			}
		}
		q := callback.Query()
		q.Set("fritzpayID", paymentTx.FritzpayID.String)
		q.Set("status", paymentTx.Status)
		callback.RawQuery = q.Encode()
		req, err = http.NewRequest("GET", callback.String(), nil)
		if err != nil {
			errors <- err
			return
		}
		res, err := cl.Do(req)
		if err != nil {
			errors <- err
			return
		}
		if res.StatusCode != http.StatusOK {
			paymentTx.Timestamp = time.Now()
			paymentTx.Status = TransactionPSPError
			paymentTx.Payload.String = "error reaching callback URL"
			err = InsertPaymentTransactionTx(tx, paymentTx)
			if err != nil {
				errors <- err
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Crit("error on commit", log15.Ctx{"err": err})
			errors <- err
			return
		}

		close(ok)
	}()
	select {
	case <-ctx.Done():
		log.Warn("cancelling worker...", log15.Ctx{"err": ctx.Err()})
		err = tx.Rollback()
		if err != nil {
			log.Crit("error on rollback", log15.Ctx{"err": err})
		}
		tr.CancelRequest(req)
		return
	case err := <-errors:
		log.Error("error on worker", log15.Ctx{"err": err})
		err = tx.Rollback()
		if err != nil {
			log.Crit("error on rollback", log15.Ctx{"err": err})
		}
		return
	case <-ok:
		if Debug {
			log.Debug("worker done")
		}
		return
	}
}

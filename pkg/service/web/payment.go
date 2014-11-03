package web

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/go-sql-driver/mysql"
	"gopkg.in/inconshreveable/log15.v2"
	"hash"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	PaymentTokenParam        = "token"
	PaymentCookieName        = "payment"
	PaymentCookieMaxLifetime = 15 * time.Minute
	PaymentAuthPaymentID     = "paymentID"
)

func (h *Handler) hashFunc() func() hash.Hash {
	return sha256.New
}

func (h *Handler) authenticatePaymentRequest(w http.ResponseWriter, r *http.Request) (proceed bool) {
	// if token present
	if tokenStr := r.URL.Query().Get(PaymentTokenParam); tokenStr != "" {
		h.authenticatePaymentToken(w, r, tokenStr)
		return false
	}
	// payment auth must be in cookie
	return h.readPaymentCookie(w, r)
}

func (h *Handler) authenticatePaymentToken(w http.ResponseWriter, r *http.Request, tokenStr string) {
	log := h.log.New(log15.Ctx{"method": "authenticatePaymentToken"})

	var tx *sql.Tx
	var commit bool
	var err error
	defer func() {
		if tx != nil && !commit {
			err = tx.Rollback()
			if err != nil {
				log.Crit("error on rollback", log15.Ctx{"err": err})
			}
		}
	}()
	maxRetries := h.ctx.Config().Database.TransactionMaxRetries
	var retries int
beginTx:
	if retries >= maxRetries {
		// no need to roll back
		commit = true
		log.Crit("too many retries on tx. aborting...", log15.Ctx{"maxRetries": maxRetries})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tx, err = h.ctx.PaymentDB(service.ReadOnly).Begin()
	if err != nil {
		commit = true
		log.Crit("error on begin tx", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	p, err := h.paymentService.PaymentByToken(tx, tokenStr)
	if err != nil {
		if err == payment.ErrPaymentNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Error("error retrieving payment token", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !p.Valid() {
		log.Crit("received invalid payment")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = h.paymentService.DeletePaymentToken(tx, tokenStr)
	if err != nil {
		if err == paymentService.ErrDBLockTimeout {
			retries++
			time.Sleep(time.Second)
			goto beginTx
		}
		log.Error("error deleting payment token", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = h.setPaymentCookie(w, p)
	if err != nil {
		log.Error("error setting cookie", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = tx.Commit()
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1213 {
				retries++
				time.Sleep(time.Second)
				goto beginTx
			}
		}
		log.Crit("error on commit", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	commit = true
	// remove query from current request, redirect
	redirectURL := &(*r.URL)
	redirectURL.RawQuery = ""
	http.Redirect(w, r, redirectURL.String(), http.StatusMovedPermanently)
}

func (h *Handler) readPaymentCookie(w http.ResponseWriter, r *http.Request) (proceed bool) {
	log := h.log.New(log15.Ctx{"method": "readPaymentCookie"})
	if c, err := r.Cookie(PaymentCookieName); err == nil {
		auth := service.NewAuthorization(h.hashFunc())
		_, err = auth.ReadFrom(strings.NewReader(c.Value))
		if err != nil {
			if err == io.EOF {
				w.WriteHeader(http.StatusNotFound)
				return false
			}
			log.Warn("error reading cookie auth", log15.Ctx{"err": err})
			h.resetPaymentCookie(w)
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		if auth.Expiry().Before(time.Now()) {
			log.Warn("expired cookie", log15.Ctx{"expiry": auth.Expiry()})
			h.resetPaymentCookie(w)
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		key, err := h.ctx.WebKeychain().MatchKey(auth)
		if err != nil {
			log.Warn("error retrieving key", log15.Ctx{"err": err})
			h.resetPaymentCookie(w)
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		err = auth.Decode(key)
		if err != nil {
			log.Error("error decoding auth", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return false
		}
		paymentIDStr, ok := auth.Payload[PaymentAuthPaymentID].(string)
		if !ok {
			log.Crit("payload type error", log15.Ctx{"hasType": fmt.Sprintf("%T", auth.Payload[PaymentAuthPaymentID])})
			w.WriteHeader(http.StatusInternalServerError)
			return false
		}
		service.SetRequestContextVar(r, PaymentAuthPaymentID, paymentIDStr)
		return true
	}
	// no cookie set
	w.WriteHeader(http.StatusNotFound)
	return false
}

func (h *Handler) setPaymentCookie(w http.ResponseWriter, p *payment.Payment) error {
	log := h.log.New(log15.Ctx{"method": "setPaymentCookie"})

	auth := service.NewAuthorization(h.hashFunc())
	auth.Payload[PaymentAuthPaymentID] = p.PaymentID().String()
	auth.Expires(time.Now().Add(PaymentCookieMaxLifetime))
	key, err := h.ctx.WebKeychain().BinKey()
	if err != nil {
		log.Error("error retrieving auth key", log15.Ctx{"err": err})
		return err
	}
	err = auth.Encode(key)
	if err != nil {
		log.Error("error encoding auth", log15.Ctx{"err": err})
		return err
	}
	c := &http.Cookie{
		Name:    PaymentCookieName,
		Path:    PaymentPath,
		Expires: auth.Expiry(),
	}
	c.Value, err = auth.Serialized()
	if err != nil {
		log.Error("error retrieving serialized auth", log15.Ctx{"err": err})
		return err
	}
	c.HttpOnly = h.ctx.Config().Web.Cookie.HTTPOnly
	c.Secure = h.ctx.Config().Web.Cookie.Secure

	http.SetCookie(w, c)
	return nil
}

func (h *Handler) resetPaymentCookie(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:    PaymentCookieName,
		Path:    PaymentPath,
		Value:   "",
		Expires: time.Unix(0, 0),
	}
	c.HttpOnly = h.ctx.Config().Web.Cookie.HTTPOnly
	c.Secure = h.ctx.Config().Web.Cookie.Secure
	http.SetCookie(w, c)
}

func (h *Handler) PaymentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		log := h.log.New(log15.Ctx{"method": "PaymentHandler"})
		// will set the appropriate header if false
		if !h.authenticatePaymentRequest(w, r) {
			return
		}
		paymentIDStr, ok := service.RequestContext(r).Value(PaymentAuthPaymentID).(string)
		if !ok {
			log.Crit("error in request context payment id", log15.Ctx{"hasType": fmt.Sprintf("%T", service.RequestContext(r).Value(PaymentAuthPaymentID))})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		paymentID, err := payment.ParsePaymentIDStr(paymentIDStr)
		if err != nil {
			log.Crit("invalid payment id", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log = log.New(log15.Ctx{
			"displayPaymentId": h.paymentService.EncodedPaymentID(paymentID).String(),
			"projectID":        paymentID.ProjectID,
			"paymentID":        paymentID.PaymentID,
		})
		if Debug {
			log.Debug("handling payment...")
		}
	})
}

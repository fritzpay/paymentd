package web

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/text/language"
	"gopkg.in/inconshreveable/log15.v2"
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
	h.redirectTokenRequest(w, r)
}

func (h *Handler) redirectTokenRequest(w http.ResponseWriter, r *http.Request) {
	// remove query from current request, redirect
	redirectURL := &(*r.URL)
	q := r.URL.Query()
	q.Del("token")
	redirectURL.RawQuery = q.Encode()
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
	c.Secure = h.ctx.Config().Web.Secure

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
	c.Secure = h.ctx.Config().Web.Secure
	http.SetCookie(w, c)
}

func (h *Handler) PaymentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// will set the appropriate header if false
		if !h.authenticatePaymentRequest(w, r) {
			return
		}
		log := h.log.New(log15.Ctx{"method": "PaymentHandler"})
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
		})

		var tx *sql.Tx
		var commit bool
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
		tx, err = h.ctx.PaymentDB().Begin()
		if err != nil {
			commit = true
			log.Crit("error on begin tx", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		p, err := payment.PaymentByIDTx(tx, paymentID)
		if err != nil {
			if err == payment.ErrPaymentNotFound {
				log.Warn("requested payment not found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
			log.Error("error retrieving payment", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log = log.New(log15.Ctx{
			"projectID": p.ProjectID(),
			"paymentID": p.ID(),
		})
		if Debug {
			log.Debug("handling payment...")
		}

		var configChanged, metadataChanged bool
		h.determineLocale(p, r, &configChanged, &metadataChanged)
		h.determineEnv(p, r, &configChanged, &metadataChanged)
		var method *payment_method.Method
		method, err = h.determinePaymentMethodID(tx, p, w, r, &configChanged, &metadataChanged)
		if err != nil {
			log.Warn("error determining payment method id", log15.Ctx{"err": err})
			return
		}

		if configChanged {
			err = h.paymentService.SetPaymentConfig(tx, p)
			if err != nil {
				if err == paymentService.ErrDBLockTimeout {
					retries++
					time.Sleep(time.Second)
					goto beginTx
				}
				log.Error("error on saving payment config", log15.Ctx{"err": err})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		if metadataChanged {
			err = h.paymentService.SetPaymentMetadata(tx, p)
			if err != nil {
				if err == paymentService.ErrDBLockTimeout {
					retries++
					time.Sleep(time.Second)
					goto beginTx
				}
				log.Error("error on saving payment metadata", log15.Ctx{"err": err})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// select payment method id?
		// TODO depending on configuration this might not be wanted
		// payment method id selection fallback?
		if !p.Config.PaymentMethodID.Valid {
			if Debug {
				log.Debug("will serve payment method selection...")
			}
			err = tx.Commit()
			if err != nil {
				if mysqlErr, ok := err.(*mysql.MySQLError); ok {
					// lock error
					if mysqlErr.Number == 1213 {
						retries++
						time.Sleep(time.Second)
						goto beginTx
					}
				}
				commit = true
				log.Crit("error on commit tx", log15.Ctx{"err": err})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			commit = true
			h.SelectPaymentMethodHandler(p).ServeHTTP(w, r)
			return
		}
		// cannot process payment with the information we collected
		if !h.paymentService.IsProcessablePayment(p) {
			log.Error("payment requested but not processable. not recoverable")
			w.WriteHeader(http.StatusConflict)
			return
		}
		var paymentTx *payment.PaymentTransaction
		// payment is not initialized, set open status
		if !h.paymentService.IsInitialized(p) {
			// open transaction, ledger is -1 * amount (open payment has negative balance)
			paymentTx = p.NewTransaction(payment.PaymentStatusOpen)
			paymentTx.Amount *= -1
			err = h.paymentService.SetPaymentTransaction(tx, paymentTx)
			if err != nil {
				if err == paymentService.ErrDBLockTimeout {
					retries++
					time.Sleep(time.Second)
					goto beginTx
				}
				log.Error("error on saving payment transaction", log15.Ctx{"err": err})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				// lock error
				if mysqlErr.Number == 1213 {
					retries++
					time.Sleep(time.Second)
					goto beginTx
				}
			}
			commit = true
			log.Crit("error on commit tx", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		commit = true

		// do callback notification when a new payment transaction was created
		if paymentTx != nil {
			h.paymentService.CallbackPaymentTransaction(paymentTx)
		}

		h.servePaymentHandler(p, method).ServeHTTP(w, r)
	})
}

func (h *Handler) determineLocale(p *payment.Payment, r *http.Request, configChanged, metadataChanged *bool) {
	if p.Config.Locale.Valid {
		return
	}
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	acceptLang := r.Header.Get("Accept-Language")
	if _, ok := p.Metadata[payment.MetadataKeyAcceptLanguage]; !ok {
		if acceptLang != "" {
			p.Metadata[payment.MetadataKeyAcceptLanguage] = acceptLang
			*metadataChanged = true
		}
	}
	var locale string
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err == nil && len(tags) >= 1 {
		locale = tags[0].String()
		p.Metadata[payment.MetadataKeyBrowserLocale] = locale
		*metadataChanged = true
	}
	if locale == "" {
		locale = payment.DefaultLocale
	}
	p.Config.SetLocale(locale)
	*configChanged = true
	return
}

func (h *Handler) determineEnv(p *payment.Payment, r *http.Request, configChanged, metadataChanged *bool) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	if _, ok := p.Metadata[payment.MetadataKeyRemoteAddress]; !ok {
		p.Metadata[payment.MetadataKeyRemoteAddress] = r.RemoteAddr
		*metadataChanged = true
	}
}

func (h *Handler) determinePaymentMethodID(tx *sql.Tx, p *payment.Payment, w http.ResponseWriter, r *http.Request, configChanged, metadataChanged *bool) (*payment_method.Method, error) {
	var paymentMethodID int64
	if p.Config.PaymentMethodID.Valid {
		paymentMethodID = p.Config.PaymentMethodID.Int64
	} else if idStr := r.URL.Query().Get("paymentMethodId"); idStr != "" {
		var err error
		paymentMethodID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, fmt.Errorf("invalid payment method id: %s", idStr)
		}
	}
	meth, err := payment_method.PaymentMethodByIDTx(tx, paymentMethodID)
	if err != nil {
		if err == payment_method.ErrPaymentMethodNotFound {
			w.WriteHeader(http.StatusNotFound)
			return nil, fmt.Errorf("payment method id %d not found", paymentMethodID)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("error selecting payment method id: %v", err)
	}
	if meth.ProjectID != p.ProjectID() {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("invalid payment method id %d. project mismatch", paymentMethodID)
	}
	if !meth.Active() {
		w.WriteHeader(http.StatusConflict)
		return nil, fmt.Errorf("invalid payment method id %d. payment method not active", paymentMethodID)
	}
	if !p.Config.PaymentMethodID.Valid {
		p.Config.SetPaymentMethodID(meth.ID)
		*configChanged = true
	}
	return meth, nil
}

func (h *Handler) servePaymentHandler(p *payment.Payment, method *payment_method.Method) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := h.log.New(log15.Ctx{
			"method":          "servePaymentHandler",
			"projectID":       p.ProjectID(),
			"paymentID":       p.ID(),
			"paymentMethodID": method.ID,
			"providerID":      method.Provider.ID,
		})
		driver, err := h.providerService.Driver(method)
		if err != nil {
			log.Crit("error retrieving driver", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if Debug {
			log.Debug("initializing payment with driver...")
		}
		h, err := driver.InitPayment(p, method)
		if err != nil {
			log.Error("error on driver init payment", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.ServeHTTP(w, r)
	})
}

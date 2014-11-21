package payment

import (
	"database/sql"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/server"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/go-sql-driver/mysql"
	"gopkg.in/inconshreveable/log15.v2"
)

type errorID int

func (e errorID) Error() string {
	switch e {
	case ErrDB:
		return "database error"
	case ErrDBLockTimeout:
		return "lock wait timeout"
	case ErrDuplicateIdent:
		return "duplicate ident in payment"
	case ErrPaymentCallbackConfig:
		return "callback config error"
	case ErrPaymentMethodNotFound:
		return "payment method not found"
	case ErrPaymentMethodConflict:
		return "payment method project mismatch"
	case ErrPaymentMethodInactive:
		return "payment method inactive"
	case ErrInternal:
		return "internal error"
	case ErrIntentTimeout:
		return "intent timeout"
	default:
		return "unknown error"
	}
}

const (
	// general database error
	ErrDB errorID = iota
	// lock wait timeout
	ErrDBLockTimeout
	// duplicate Ident in payment
	ErrDuplicateIdent
	// callback config error
	ErrPaymentCallbackConfig
	// payment method not found
	ErrPaymentMethodNotFound
	// payment method project mismatch
	ErrPaymentMethodConflict
	// payment method inactive
	ErrPaymentMethodInactive
	// internal error
	ErrInternal
	// intent timeout
	ErrIntentTimeout
)

const (
	notificationBufferSize = 16
)

const (
	// PaymentTokenMaxAgeDefault is the default maximum age of payment tokens
	PaymentTokenMaxAgeDefault = time.Minute * 15
)

// IntentWorker is the primary means of synchronizing and controlling changes on payment
// states.
//
// IntentWorkers are registered with the payment service via the Service.RegiserIntentWorker
// method.
//
// Whenever another service or process wishes to change the state of a payment, it should
// do so by invoking one of the Intent* methods. These methods will create the
// matching PaymentTransaction types and start the intent procedure.
//
// PreIntent is invoked prior to the intent creation. Any errors sent through the res channel
// will cancel the intent procedure and the calling service will receive the first
// encountered error. Once the done channel is closed, the intent procedure won't accept any
// results of the IntentWorker anymore. This is usually due to timeout.
//
// PostIntent is invoked concurrently right before the Intent* methods will return the
// matching Transaction. At this point the intent cannot be cancelled. Any errors sent
// through the returned channel will be logged.
type IntentWorker interface {
	PreIntent(p payment.Payment, paymentTx payment.PaymentTransaction, done <-chan struct{}, res chan<- error)
	PostIntent(p payment.Payment, paymentTx payment.PaymentTransaction) <-chan error
}

// Service is the payment service
type Service struct {
	ctx *service.Context
	log log15.Logger

	idCoder *payment.IDEncoder

	Notify chan *payment.PaymentTransaction

	tr *http.Transport
	cl *http.Client

	mIntent sync.RWMutex
	intents map[string][]IntentWorker
}

// NewService creates a new payment service
func NewService(ctx *service.Context) (*Service, error) {
	s := &Service{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/payment",
		}),

		intents: make(map[string][]IntentWorker),
	}

	var err error
	cfg := ctx.Config()

	s.idCoder, err = payment.NewIDEncoder(cfg.Payment.PaymentIDEncPrime, cfg.Payment.PaymentIDEncXOR)
	if err != nil {
		s.log.Error("error initializing payment ID encoder", log15.Ctx{"err": err})
		return nil, err
	}

	s.Notify = make(chan *payment.PaymentTransaction, notificationBufferSize)

	s.tr = &http.Transport{}
	s.cl = &http.Client{
		Transport: s.tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 {
				return errors.New("too many redirects")
			}
			// keep user-agent
			if len(via) > 0 {
				lastReq := via[len(via)-1]
				if lastReq.Header.Get("User-Agent") != "" {
					req.Header.Set("User-Agent", lastReq.Header.Get("User-Agent"))
				}
			}
			return nil
		},
	}

	go s.handleBackground()

	return s, nil
}

func (s *Service) handleBackground() {
	// if attached to a server, this will tell the server to wait with shutting down
	// until the cleanup process is complete
	server.Wait.Add(1)
	defer server.Wait.Done()
	for {
		select {
		case <-s.ctx.Done():
			s.log.Info("service context closed", log15.Ctx{"err": s.ctx.Err()})
			s.log.Info("closing idle connections...")
			s.tr.CloseIdleConnections()
			return

		case paymentTx := <-s.Notify:
			if paymentTx == nil {
				break
			}
			err := s.notify(paymentTx)
			if err != nil {
				s.log.Error("error on callback", log15.Ctx{"err": err})
			}
		}
	}
}

func (s *Service) RegisterIntentWorker(intent string, worker IntentWorker) {
	s.mIntent.Lock()
	if s.intents[intent] == nil {
		s.intents[intent] = make([]IntentWorker, 0, 16)
	}
	s.intents[intent] = append(s.intents[intent], worker)
	s.mIntent.Unlock()
}

// EncodedPaymentID returns a payment id with the id part encoded
func (s *Service) EncodedPaymentID(id payment.PaymentID) payment.PaymentID {
	id.PaymentID = s.idCoder.Hide(id.PaymentID)
	return id
}

// DecodedPaymentID returns a payment id with the id part decoded
func (s *Service) DecodedPaymentID(id payment.PaymentID) payment.PaymentID {
	id.PaymentID = s.idCoder.Show(id.PaymentID)
	return id
}

// CreatePayment creates a new payment
func (s *Service) CreatePayment(tx *sql.Tx, p *payment.Payment) error {
	log := s.log.New(log15.Ctx{
		"method": "CreatePayment",
	})
	if p.Config.HasCallback() {
		callbackProjectKey, err := project.ProjectKeyByKeyDB(s.ctx.PrincipalDB(service.ReadOnly), p.Config.CallbackProjectKey.String)
		if err != nil {
			if err == project.ErrProjectKeyNotFound {
				log.Error("callback project key not found", log15.Ctx{"callbackProjectKey": p.Config.CallbackProjectKey.String})
				return ErrPaymentCallbackConfig
			}
			log.Error("error retrieving callback project key", log15.Ctx{"err": err})
			return ErrDB
		}
		if callbackProjectKey.Project.ID != p.ProjectID() {
			log.Error("callback project mismatch", log15.Ctx{
				"callbackProjectKey": callbackProjectKey.Key,
				"callbackProjectID":  callbackProjectKey.Project.ID,
				"projectID":          p.ProjectID(),
			})
			return ErrPaymentCallbackConfig
		}
	}
	err := payment.InsertPaymentTx(tx, p)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1213 {
				return ErrDBLockTimeout
			}
		}
		_, existErr := payment.PaymentByProjectIDAndIdentTx(tx, p.ProjectID(), p.Ident)
		if existErr != nil && existErr != payment.ErrPaymentNotFound {
			log.Error("error on checking duplicate ident", log15.Ctx{"err": err})
			return ErrDB
		}
		// payment found => duplicate error
		if existErr == nil {
			return ErrDuplicateIdent
		}
		log.Error("error on insert payment", log15.Ctx{"err": err})
		return ErrDB
	}
	err = s.SetPaymentConfig(tx, p)
	if err != nil {
		return err
	}
	err = s.SetPaymentMetadata(tx, p)
	if err != nil {
		return err
	}
	return nil
}

// SetPaymentConfig sets/updates the payment configuration
func (s *Service) SetPaymentConfig(tx *sql.Tx, p *payment.Payment) error {
	log := s.log.New(log15.Ctx{"method": "SetPaymentConfig"})
	if p.Config.PaymentMethodID.Valid {
		log = log.New(log15.Ctx{"paymentMethodID": p.Config.PaymentMethodID.Int64})
		meth, err := payment_method.PaymentMethodByIDTx(tx, p.Config.PaymentMethodID.Int64)
		if err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				if mysqlErr.Number == 1213 {
					return ErrDBLockTimeout
				}
			}
			if err == payment_method.ErrPaymentMethodNotFound {
				log.Warn(ErrPaymentMethodNotFound.Error())
				return ErrPaymentMethodNotFound
			}
			log.Error("error on select payment method", log15.Ctx{"err": err})
			return ErrDB
		}
		if meth.ProjectID != p.ProjectID() {
			log.Warn(ErrPaymentMethodConflict.Error())
			return ErrPaymentMethodConflict
		}
		if !meth.Active() {
			log.Warn(ErrPaymentMethodInactive.Error())
			return ErrPaymentMethodInactive
		}
	}
	err := payment.InsertPaymentConfigTx(tx, p)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1213 {
				return ErrDBLockTimeout
			}
		}
		log.Error("error on insert payment config", log15.Ctx{"err": err})
		return ErrDB
	}
	return nil
}

// SetPaymentMetadata sets/updates the payment metadata
func (s *Service) SetPaymentMetadata(tx *sql.Tx, p *payment.Payment) error {
	log := s.log.New(log15.Ctx{"method": "SetPaymentMetadata"})
	// payment metadata
	if p.Metadata == nil {
		return nil
	}
	err := payment.InsertPaymentMetadataTx(tx, p)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1213 {
				return ErrDBLockTimeout
			}
		}
		log.Error("error on insert payment metadata", log15.Ctx{"err": err})
		return ErrDB
	}
	return nil
}

// IsProcessablePayment returns true if the given payment is considered processable
//
// All required fields are present.
func (s *Service) IsProcessablePayment(p *payment.Payment) bool {
	if !p.Config.IsConfigured() {
		return false
	}
	if !p.Config.Country.Valid {
		return false
	}
	if !p.Config.Locale.Valid {
		return false
	}
	if !p.Config.PaymentMethodID.Valid {
		return false
	}
	return true
}

// IsInitialized returns true when the payment is in a processing state, i.e.
// when there is at least one transaction present
func (s *Service) IsInitialized(p *payment.Payment) bool {
	return p.Status != payment.PaymentStatusNone
}

// SetPaymentTransaction adds a new payment transaction
//
// If a callback method is configured for this payment/project, it will send a callback
// notification
func (s *Service) SetPaymentTransaction(tx *sql.Tx, paymentTx *payment.PaymentTransaction) error {
	log := s.log.New(log15.Ctx{"method": "SetPaymentTransaction"})
	err := payment.InsertPaymentTransactionTx(tx, paymentTx)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1213 {
				return ErrDBLockTimeout
			}
		}
		log.Error("error saving payment transaction", log15.Ctx{"err": err})
		return ErrDB
	}
	return nil
}

// PaymentTransaction returns the current payment transaction for the given payment
//
// PaymentTransaction will return a payment.ErrPaymentTransactionNotFound if no such
// transaction exists (i.e. the payment is uninitialized)
func (s *Service) PaymentTransaction(tx *sql.Tx, p *payment.Payment) (*payment.PaymentTransaction, error) {
	return payment.PaymentTransactionCurrentTx(tx, p)
}

func (s *Service) IntentOpen(p *payment.Payment, timeout time.Duration) (*payment.PaymentTransaction, error) {
	if deadline, ok := s.ctx.Deadline(); ok {
		if time.Now().Add(timeout).After(deadline) {
			return nil, ErrIntentTimeout
		}
	}

	paymentTx := p.NewTransaction(payment.PaymentStatusOpen)
	paymentTx.Amount = paymentTx.Amount * -1

	s.mIntent.RLock()
	if len(s.intents["open"]) > 0 {
		done := make(chan struct{})
		c := make(chan error, 1)
		for _, w := range s.intents["open"] {
			go w.PreIntent(*p, *paymentTx, done, c)
		}
		select {
		case <-s.ctx.Done():
			close(done)
			s.mIntent.RUnlock()
			return nil, s.ctx.Err()

		case err := <-c:
			close(done)
			s.mIntent.RUnlock()
			return nil, err

		case <-time.After(timeout):
			close(done)
		}

		postDone := make([]<-chan error, len(s.intents["open"]))
		for i, w := range s.intents["open"] {
			postDone[i] = w.PostIntent(*p, *paymentTx)
		}
		waitFunc := func(wait []<-chan error) {
			var wg sync.WaitGroup
			for _, w := range wait {
				wg.Add(1)
				go func(c <-chan error) {
					err, ok := <-c
					if ok && err != nil {
						s.log.Warn("error on post intent action", log15.Ctx{
							"intent": "open",
							"err":    err,
						})
					}
					wg.Done()
				}(w)
			}
			wg.Wait()
		}
		go waitFunc(postDone)
	}
	s.mIntent.RUnlock()

	return paymentTx, nil
}

// CreatePaymentToken creates a new random payment token
func (s *Service) CreatePaymentToken(tx *sql.Tx, p *payment.Payment) (*payment.PaymentToken, error) {
	log := s.log.New(log15.Ctx{"method": "CreatePaymentToken"})
	token, err := payment.NewPaymentToken(p.PaymentID())
	if err != nil {
		log.Error("error creating payment token", log15.Ctx{"err": err})
		return nil, ErrInternal
	}
	err = payment.InsertPaymentTokenTx(tx, token)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1213 {
				return nil, ErrDBLockTimeout
			}
		}
		log.Error("error saving payment token", log15.Ctx{"err": err})
		return nil, ErrDB
	}
	return token, nil
}

// PaymentByToken returns the payment associated with the given payment token
//
// TODO use token max age from config
func (s *Service) PaymentByToken(tx *sql.Tx, token string) (*payment.Payment, error) {
	tokenMaxAge := PaymentTokenMaxAgeDefault
	return payment.PaymentByTokenTx(tx, token, tokenMaxAge)
}

// DeletePaymentToken deletes/invalidates the given payment token
func (s *Service) DeletePaymentToken(tx *sql.Tx, token string) error {
	log := s.log.New(log15.Ctx{"method": "DeletePaymentToken"})
	err := payment.DeletePaymentTokenTx(tx, token)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1213 {
				return ErrDBLockTimeout
			}
		}
		log.Error("error deleting payment token", log15.Ctx{"err": err})
		return ErrDB
	}
	return nil
}

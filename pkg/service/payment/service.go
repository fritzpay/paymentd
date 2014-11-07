package payment

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/server"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/go-sql-driver/mysql"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"time"
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
)

const (
	// PaymentTokenMaxAgeDefault is the default maximum age of payment tokens
	PaymentTokenMaxAgeDefault = time.Minute * 15
)

// Service is the payment service
type Service struct {
	ctx *service.Context
	log log15.Logger

	idCoder *payment.IDEncoder

	tr *http.Transport
	cl *http.Client
}

// NewService creates a new payment service
func NewService(ctx *service.Context) (*Service, error) {
	s := &Service{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/payment",
		}),
	}

	var err error
	cfg := ctx.Config()

	s.idCoder, err = payment.NewIDEncoder(cfg.Payment.PaymentIDEncPrime, cfg.Payment.PaymentIDEncXOR)
	if err != nil {
		s.log.Error("error initializing payment ID encoder", log15.Ctx{"err": err})
		return nil, err
	}

	s.tr = &http.Transport{}
	s.cl = &http.Client{
		Transport: s.tr,
	}

	go s.handleContext()

	return s, nil
}

func (s *Service) handleContext() {
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
		}
	}
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
		callbackProjectKey, err := project.ProjectKeyByKeyTx(tx, p.Config.CallbackProjectKey.String)
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
		if meth.Status != payment_method.PaymentMethodStatusActive {
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
	err = s.CallbackPaymentTransaction(tx, paymentTx)
	if err != nil {
		return err
	}
	return nil
}

// CallbackPaymentTransaction performs a callback notification if the payment/project has
// a callback configured
func (s *Service) CallbackPaymentTransaction(tx *sql.Tx, paymentTx *payment.PaymentTransaction) error {
	log := s.log.New(log15.Ctx{"method": "CallbackPaymentTransaction"})
	var callback Callbacker
	if CanCallback(&paymentTx.Payment.Config) {
		callback = &paymentTx.Payment.Config
	} else {
		pr, err := project.ProjectByIDTx(tx, paymentTx.Payment.ProjectID())
		if err != nil {
			if err == project.ErrProjectNotFound {
				log.Crit("payment with invalid project", log15.Ctx{"projectID": paymentTx.Payment.ProjectID()})
				return ErrInternal
			}
			log.Error("error retrieving project", log15.Ctx{"err": err})
			return ErrDB
		}
		if CanCallback(pr.Config) {
			callback = pr.Config
		}
	}
	if callback != nil {
		s.Notify(callback, paymentTx)
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

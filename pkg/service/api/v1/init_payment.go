package v1

import (
	"bytes"
	"code.google.com/p/go.text/language"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	jsonutil "github.com/fritzpay/paymentd/pkg/json"
	"github.com/fritzpay/paymentd/pkg/maputil"
	"github.com/fritzpay/paymentd/pkg/paymentd/currency"
	"github.com/fritzpay/paymentd/pkg/paymentd/nonce"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/go-sql-driver/mysql"
	"gopkg.in/inconshreveable/log15.v2"
	"hash"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode/utf8"
)

const (
	initPaymentTimestampMaxAge = time.Minute
)

// InitPaymentRequest is the request JSON struct for POST /payment
type InitPaymentRequest struct {
	ProjectKey      string
	Ident           string
	Amount          jsonutil.RequiredInt64
	Subunits        jsonutil.RequiredInt8
	Currency        string
	Country         string
	PaymentMethodID int64  `json:"PaymentMethodId,string"`
	Locale          string `json:",omitempty"`
	CallbackURL     string `json:",omitempty"`
	ReturnURL       string `json:",omitempty"`

	Metadata map[string]string

	Timestamp int64 `json:",string"`
	Nonce     string

	HexSignature    string `json:"Signature"`
	binarySignature []byte
}

// Validate input
func (r *InitPaymentRequest) Validate() error {
	if r.ProjectKey == "" {
		return fmt.Errorf("missing ProjectKey")
	}
	if r.Ident == "" {
		return fmt.Errorf("missing Ident")
	}
	if utf8.RuneCountInString(r.Ident) > payment.IdentMaxLen {
		return fmt.Errorf("invalid Ident")
	}
	if !r.Amount.Set {
		return fmt.Errorf("missing Amount")
	}
	if r.Amount.Int64 < 0 {
		return fmt.Errorf("invalid Amount: %d", r.Amount.Int64)
	}
	if !r.Subunits.Set {
		return fmt.Errorf("missing Subunits")
	}
	if r.Currency == "" {
		return fmt.Errorf("missing Currency")
	}
	if len(r.Currency) != 3 {
		return fmt.Errorf("invalid Currency")
	}
	if r.Country == "" {
		return fmt.Errorf("missing Country")
	}
	if len(r.Country) != 2 {
		return fmt.Errorf("invalid Country")
	}
	if r.Timestamp == 0 {
		return fmt.Errorf("missing Timestamp")
	}
	if r.Nonce == "" {
		return fmt.Errorf("missing Nonce")
	}
	if len(r.Nonce) > nonce.NonceBytes {
		return fmt.Errorf("invalid Nonce")
	}
	var err error
	if r.HexSignature == "" {
		return fmt.Errorf("missing Signature")
	} else if r.binarySignature, err = hex.DecodeString(r.HexSignature); err != nil {
		return fmt.Errorf("invalid Signature format")
	}
	if r.Locale != "" {
		if _, err := language.Parse(r.Locale); err != nil {
			return fmt.Errorf("invalid Locale")
		}
	}
	if r.CallbackURL != "" {
		if _, err = url.Parse(r.CallbackURL); err != nil {
			return fmt.Errorf("invalid CallbackURL")
		}
	}
	if r.ReturnURL != "" {
		if _, err := url.Parse(r.ReturnURL); err != nil {
			return fmt.Errorf("invalid ReturnURL")
		}
	}
	return nil
}

// Return the (binary) signature from the request
//
// implementing AuthenticatedRequest
func (r *InitPaymentRequest) Signature() []byte {
	return r.binarySignature
}

// Message returns the signature base string as bytes or nil on error
func (r *InitPaymentRequest) Message() []byte {
	str, err := r.SignatureBaseString()
	if err != nil {
		return nil
	}
	return []byte(str)
}

// HashFunc returns the hash function used to generate a signature
func (r *InitPaymentRequest) HashFunc() func() hash.Hash {
	return sha256.New
}

// Return the signature base string (msg)
func (r *InitPaymentRequest) SignatureBaseString() (string, error) {
	var err error
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(r.ProjectKey)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Ident)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Amount.Int64, 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(int64(r.Subunits.Int8), 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Currency)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Country)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	if r.PaymentMethodID != 0 {
		_, err = buf.WriteString(strconv.FormatInt(r.PaymentMethodID, 10))
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.Locale != "" {
		_, err = buf.WriteString(r.Locale)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.CallbackURL != "" {
		_, err = buf.WriteString(r.CallbackURL)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.ReturnURL != "" {
		_, err = buf.WriteString(r.ReturnURL)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.Metadata != nil {
		err = maputil.WriteSortedMap(buf, r.Metadata)
		if err != nil {
			return "", fmt.Errorf("error writing map: %v", err)
		}
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Timestamp, 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Nonce)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	s := buf.String()
	return s, nil
}

func (r *InitPaymentRequest) ReadJSON(rd io.Reader) error {
	dec := json.NewDecoder(rd)
	err := dec.Decode(r)
	return err
}

func (r *InitPaymentRequest) PopulatePaymentFields(p *payment.Payment) {
	p.Ident = r.Ident
	p.Amount = r.Amount.Int64
	p.Subunits = r.Subunits.Int8
	// payment config fields
	if r.PaymentMethodID != 0 {
		p.Config.SetPaymentMethodID(r.PaymentMethodID)
	}
	if r.Country != "" {
		p.Config.SetCountry(r.Country)
	}
	if r.Locale != "" {
		p.Config.SetLocale(r.Locale)
	}
	if r.CallbackURL != "" {
		p.Config.SetCallbackURL(r.CallbackURL)
	}
	if r.ReturnURL != "" {
		p.Config.SetReturnURL(r.ReturnURL)
	}
	if r.Metadata != nil {
		p.Metadata = r.Metadata
	}
}

// InitPaymentResponse is the JSON response struct for POST /payment
type InitPaymentResponse struct {
	Confirmation struct {
		Ident           string
		Amount          int64 `json:",string"`
		Subunits        int8  `json:",string"`
		Currency        string
		Country         string
		PaymentMethodID int64             `json:"PaymentMethodId,string,omitempty"`
		Locale          string            `json:",omitempty"`
		CallbackURL     string            `json:",omitempty"`
		ReturnURL       string            `json:",omitempty"`
		Metadata        map[string]string `json:",omitempty"`
	}
	Payment struct {
		PaymentId payment.PaymentID
		// RFC3339 date/time string
		Created     string
		Token       string
		RedirectURL string `json:",omitempty"`
	}
	Timestamp int64 `json:",string"`
	Nonce     string
	Signature string
}

// ConfirmationFromPayment populates the response "Confirmation" object with
// the fields from the given payment
func (r *InitPaymentResponse) ConfirmationFromPayment(p *payment.Payment) {
	r.Confirmation.Ident = p.Ident
	r.Confirmation.Amount = p.Amount
	r.Confirmation.Subunits = p.Subunits
	r.Confirmation.Currency = p.Currency

	if p.Config.Locale.Valid {
		r.Confirmation.Locale = p.Config.Locale.String
	}
	if p.Config.Country.Valid {
		r.Confirmation.Country = p.Config.Country.String
	}
	if p.Config.PaymentMethodID.Valid {
		r.Confirmation.PaymentMethodID = p.Config.PaymentMethodID.Int64
	}
	if p.Config.CallbackURL.Valid {
		r.Confirmation.CallbackURL = p.Config.CallbackURL.String
	}
	if p.Config.ReturnURL.Valid {
		r.Confirmation.ReturnURL = p.Config.ReturnURL.String
	}
	if p.Metadata != nil {
		r.Confirmation.Metadata = p.Metadata
	}
}

// Message returns the signature base string as a byte slice, nil if an error occured
func (r *InitPaymentResponse) Message() []byte {
	str, err := r.SignatureBaseString()
	if err != nil {
		return nil
	}
	return []byte(str)
}

// HashFunc returns the hash function for signing an init payment response
func (r *InitPaymentResponse) HashFunc() func() hash.Hash {
	return sha256.New
}

// Returns the signature base string
//
// implementing SignableMessage
func (r *InitPaymentResponse) SignatureBaseString() (string, error) {
	var err error
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(r.Confirmation.Ident)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Confirmation.Amount, 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(int64(r.Confirmation.Subunits), 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Confirmation.Currency)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Confirmation.Country)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	if r.Confirmation.PaymentMethodID != 0 {
		_, err = buf.WriteString(strconv.FormatInt(r.Confirmation.PaymentMethodID, 10))
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.Confirmation.Locale != "" {
		_, err = buf.WriteString(r.Confirmation.Locale)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.Confirmation.CallbackURL != "" {
		_, err = buf.WriteString(r.Confirmation.CallbackURL)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.Confirmation.ReturnURL != "" {
		_, err = buf.WriteString(r.Confirmation.ReturnURL)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.Confirmation.Metadata != nil {
		err = maputil.WriteSortedMap(buf, r.Confirmation.Metadata)
		if err != nil {
			return "", fmt.Errorf("error writing map: %v", err)
		}
	}
	_, err = buf.WriteString(r.Payment.PaymentId.String())
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Payment.Created)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Payment.Token)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Payment.RedirectURL)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Timestamp, 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Nonce)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	s := buf.String()
	return s, nil
}

func (a *PaymentAPI) authenticateMessage(projectKey *project.Projectkey, msg service.Signed) (bool, error) {
	if projectKey == nil || !projectKey.IsValid() {
		return false, fmt.Errorf("invalid project key: %+v", projectKey)
	}
	secret, err := projectKey.SecretBytes()
	if err != nil {
		return false, err
	}
	return service.IsAuthentic(msg, secret)
}

func (a *PaymentAPI) InitPayment() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		log := a.log.New(log15.Ctx{
			"method": "InitPayment",
		})
		var resp ServiceResponse
		defer func() {
			err := resp.Write(w)
			if err != nil {
				log.Error("error writing response", log15.Ctx{"err": err})
			}
		}()
		req := &InitPaymentRequest{}
		err := req.ReadJSON(r.Body)
		if err != nil {
			resp = ErrReadJson
			if Debug {
				resp.Info = err.Error()
			}
			return
		}
		err = req.Validate()
		if err != nil {
			resp = ErrInval
			resp.Info = err.Error()
			return
		}
		projectKey, err := project.ProjectKeyByKeyDB(a.ctx.PrincipalDB(service.ReadOnly), req.ProjectKey)
		if err != nil {
			if err == project.ErrProjectKeyNotFound {
				resp = ErrUnauthorized
				if Debug {
					resp.Info = fmt.Sprintf("project key %s not found", req.ProjectKey)
				}
				return
			}
			log.Error("error on retrieving project key", log15.Ctx{"err": err})
			resp = ErrDatabase
			if Debug {
				resp.Info = fmt.Sprintf("database error: %v", err)
			}
			return
		}
		if !projectKey.IsValid() {
			log.Warn("invalid project key on request", log15.Ctx{
				"ProjectKey": projectKey.Key,
			})
			resp = ErrUnauthorized
			if Debug {
				resp.Info = fmt.Sprintf("project key %s is not valid (inactive project key?)", projectKey.Key)
			}
			return
		}
		// authenticate
		// skip if dev mode
		if !Debug {
			if auth, err := a.authenticateMessage(projectKey, req); err != nil {
				log.Error("error on authenticate message", log15.Ctx{"err": err})
				resp = ErrSystem
				return
			} else if !auth {
				resp = ErrUnauthorized
				return
			}
			if time.Since(time.Unix(req.Timestamp, 0)) > initPaymentTimestampMaxAge {
				resp = ErrUnauthorized
				return
			}
			// TODO include nonce handling
		}
		// extend log info
		log = log.New(log15.Ctx{"projectId": projectKey.Project.ID})

		curr, err := currency.CurrencyByCodeISO4217DB(a.ctx.PaymentDB(service.ReadOnly), req.Currency)
		if err != nil {
			if err == currency.ErrCurrencyNotFound {
				resp = ErrInval
				resp.Info = "invalid Currency"
				return
			}
			log.Error("error retrieving currency", log15.Ctx{"err": err})
			resp = ErrDatabase
			if Debug {
				resp.Info = fmt.Sprintf("error retrieving currency: %v", err)
			}
			return
		}
		if curr.IsEmpty() {
			log.Crit("internal error. request currency is empty")
			resp = ErrSystem
			return
		}

		// create payment
		p := &payment.Payment{
			Created:  time.Now(),
			Currency: curr.CodeISO4217,
		}
		err = p.SetProject(&projectKey.Project)
		if err != nil {
			log.Error("error setting payment project", log15.Ctx{"err": err})
			resp = ErrSystem
			if Debug {
				resp.Info = fmt.Sprintf("error setting payment project: %v", err)
			}
			return
		}
		req.PopulatePaymentFields(p)

		// DB
		var tx *sql.Tx
		var commit bool
		// deferred rollback if commit == false
		defer func() {
			if tx != nil && !commit {
				txErr := tx.Rollback()
				if txErr != nil {
					log.Crit("error on rollback", log15.Ctx{"err": txErr})
					resp = ErrDatabase
					if Debug {
						resp.Info = fmt.Sprintf("error on rollback: %v", err)
					}
				}
			}
		}()
		maxRetries := a.ctx.Config().Database.TransactionMaxRetries
		var retries int
	beginTx:
		if retries >= maxRetries {
			// no need to roll back
			commit = true
			log.Crit("too many retries on tx. aborting...", log15.Ctx{"maxRetries": maxRetries})
			resp = ErrDatabase
			return
		}
		tx, err = a.ctx.PaymentDB().Begin()
		if err != nil {
			commit = true
			log.Crit("error on begin", log15.Ctx{"err": err})
			resp = ErrDatabase
			return
		}
		// actions on payment service errors
		handlePaymentServiceErr := func(err error) {
			switch err {
			case paymentService.ErrDB:
				resp = ErrDatabase
			case paymentService.ErrDuplicateIdent:
				resp = ErrConflict
				resp.Info = "your ident was already used"
			default:
				resp = ErrSystem
				log.Error("unknown error in payment service")
			}
		}

		err = a.paymentService.CreatePayment(tx, p)
		if err != nil {
			if err == paymentService.ErrDBLockTimeout {
				retries++
				time.Sleep(time.Second)
				goto beginTx
			}
			handlePaymentServiceErr(err)
			return
		}
		// payment token
		token, err := a.paymentService.CreatePaymentToken(tx, p)
		if err != nil {
			if err == paymentService.ErrDBLockTimeout {
				retries++
				time.Sleep(time.Second)
				goto beginTx
			}
			handlePaymentServiceErr(err)
			return
		}

		paymentResp := &InitPaymentResponse{}
		paymentResp.ConfirmationFromPayment(p)
		paymentResp.Payment.PaymentId = a.paymentService.EncodedPaymentID(p.PaymentID())
		paymentResp.Payment.Created = p.Created.UTC().Format(time.RFC3339)
		paymentResp.Payment.Token = token.Token

		if projectKey.Project.Config.WebURL.Valid {
			redirect, err := url.ParseRequestURI(projectKey.Project.Config.WebURL.String)
			if err != nil {
				log.Error("could not parse project URL", log15.Ctx{
					"err":    err,
					"rawURL": projectKey.Project.Config.WebURL.String,
				})
				resp = ErrSystem
				return
			}
			redirectQ := redirect.Query()
			// TODO replace token with constant which will be also used by web service
			redirectQ.Set("token", token.Token)
			redirect.RawQuery = redirectQ.Encode()
			paymentResp.Payment.RedirectURL = redirect.String()
		}

		n, err := nonce.New()
		if err != nil {
			log.Error("error generating nonce", log15.Ctx{"err": err})
			resp = ErrSystem
			return
		}
		// TODO save nonce
		paymentResp.Nonce = n.Nonce
		paymentResp.Timestamp = time.Now().Unix()

		secret, err := projectKey.SecretBytes()
		if err != nil {
			log.Error("error retrieving project secret", log15.Ctx{"err": err})
			resp = ErrSystem
			return
		}
		sig, err := service.Sign(paymentResp, secret)
		if err != nil {
			log.Error("error signing response", log15.Ctx{"err": err})
			resp = ErrSystem
			return
		}
		paymentResp.Signature = hex.EncodeToString(sig)

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
			resp = ErrDatabase
			return
		}
		commit = true

		const info = "payment initiated"
		resp.Status = StatusSuccess
		resp.Info = info
		resp.Response = paymentResp
	})
}

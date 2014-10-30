package v1

import (
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/payment"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"time"
)

const (
	requestTimestampMaxAge = 10 * time.Second
)

// API represents the payment API in the version 1.x
type PaymentAPI struct {
	ctx *service.Context
	log log15.Logger

	paymentService *payment.Service
}

// NewAPI creates a new payment API
func NewPaymentAPI(ctx *service.Context) (*PaymentAPI, error) {
	p := &PaymentAPI{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1",
			"API": "PaymentAPI",
		}),
	}
	var err error
	p.paymentService, err = payment.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type ProjectKeyRequester interface {
	service.Signed
	RequestProjectKey() string
	Time() time.Time
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

func (a *PaymentAPI) authenticateRequest(req ProjectKeyRequester, log log15.Logger, w http.ResponseWriter) *project.Projectkey {
	projectKey, err := project.ProjectKeyByKeyDB(a.ctx.PrincipalDB(service.ReadOnly), req.RequestProjectKey())
	if err != nil {
		if err == project.ErrProjectKeyNotFound {
			resp := ErrUnauthorized
			if Debug {
				resp.Info = fmt.Sprintf("project key %s not found", req.RequestProjectKey())
			}
			resp.Write(w)
			return nil
		}
		log.Error("error on retrieving project key", log15.Ctx{"err": err})
		resp := ErrDatabase
		if Debug {
			resp.Info = fmt.Sprintf("database error: %v", err)
		}
		resp.Write(w)
		return nil
	}
	if !projectKey.IsValid() {
		log.Warn("invalid project key on request", log15.Ctx{
			"ProjectKey": projectKey.Key,
		})
		resp := ErrUnauthorized
		if Debug {
			resp.Info = fmt.Sprintf("project key %s is not valid (inactive project key?)", projectKey.Key)
		}
		resp.Write(w)
		return nil
	}
	// authenticate
	// skip if dev mode
	if !Debug {
		if auth, err := a.authenticateMessage(projectKey, req); err != nil {
			log.Error("error on authenticate message", log15.Ctx{"err": err})
			ErrSystem.Write(w)
			return nil
		} else if !auth {
			ErrUnauthorized.Write(w)
			return nil
		}
		if time.Since(req.Time()) > requestTimestampMaxAge {
			ErrUnauthorized.Write(w)
			return nil
		}
		// TODO include nonce handling
	}
	return projectKey
}

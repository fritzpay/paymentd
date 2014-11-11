package paypal_rest

import (
	"database/sql"
	"errors"
	"net/url"
	"sync"

	"code.google.com/p/goauth2/oauth"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	paypalTokenPath = "/v1/oauth2/token"
)

var (
	ErrNoToken     = errors.New("no token")
	ErrNoTransport = errors.New("no transport")
)

type TokenCache struct {
	m sync.RWMutex
	t *oauth.Token
}

func NewTokenCache() *TokenCache {
	return &TokenCache{}
}

func (t *TokenCache) Token() (*oauth.Token, error) {
	t.m.RLock()
	token := t.t
	t.m.RUnlock()
	if token == nil {
		return nil, ErrNoToken
	}
	return token, nil
}

func (t *TokenCache) PutToken(token *oauth.Token) error {
	t.m.Lock()
	t.t = token
	t.m.Unlock()
	return nil
}

type OAuthTransportStore struct {
	m    sync.RWMutex
	cfgs map[int64]map[string]*oauth.Transport
}

func NewOAuthTransportStore() *OAuthTransportStore {
	return &OAuthTransportStore{cfgs: make(map[int64]map[string]*oauth.Transport)}
}

func (c *OAuthTransportStore) Transport(projectID int64, methodKey string) (*oauth.Transport, error) {
	c.m.RLock()
	if c.cfgs[projectID] == nil {
		c.m.RUnlock()
		return nil, ErrNoTransport
	}
	tr := c.cfgs[projectID][methodKey]
	c.m.RUnlock()
	if tr == nil {
		return nil, ErrNoTransport
	}
	return tr, nil
}

func (c *OAuthTransportStore) PutTransport(projectID int64, methodKey string, tr *oauth.Transport) {
	c.m.Lock()
	if c.cfgs[projectID] == nil {
		c.cfgs[projectID] = make(map[string]*oauth.Transport)
	}
	c.cfgs[projectID][methodKey] = tr
	c.m.Unlock()
}

func (d *Driver) oAuthTransport(log log15.Logger) func(tx *sql.Tx, p *payment.Payment, method *payment_method.Method) (*oauth.Transport, error) {
	return func(tx *sql.Tx, p *payment.Payment, method *payment_method.Method) (*oauth.Transport, error) {
		tr, err := d.oauth.Transport(p.ProjectID(), method.MethodKey)
		if err != nil && err != ErrNoTransport {
			log.Error("error retrieving transport", log15.Ctx{"err": err})
			return nil, ErrInternal
		}
		if err == ErrNoTransport {
			cfg, err := ConfigByPaymentMethodTx(tx, method)
			if err != nil {
				log.Error("error retrieving PayPal config", log15.Ctx{"err": err})
				return nil, ErrDatabase
			}
			tokenURL, err := url.Parse(cfg.Endpoint)
			if err != nil {
				log.Error("invalid endpoint", log15.Ctx{"err": err})
				return nil, ErrInternal
			}
			tokenURL.Path = paypalTokenPath

			oAuthCfg := &oauth.Config{
				ClientId:     cfg.ClientID,
				ClientSecret: cfg.Secret,
				TokenURL:     tokenURL.String(),
				TokenCache:   NewTokenCache(),
			}
			tr = &oauth.Transport{Config: oAuthCfg}
			d.oauth.PutTransport(p.ProjectID(), method.MethodKey, tr)
		}
		return tr, nil
	}
}

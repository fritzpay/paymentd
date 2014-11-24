package config

import (
	"encoding/json"
	"io"
	"time"
)

// DatabaseConfig represents a single database DSN
// It is a map, so the JSON representation can include the backend type and the
// DSN, e.g.
//
//   {
//     "mysql": "paymentd@tcp(localhost:3306)/fritzpay"
//   }
type DatabaseConfig map[string]string

// NewDatabaseConfig creates a new DatabaseConfig
func NewDatabaseConfig() DatabaseConfig {
	return DatabaseConfig(make(map[string]string))
}

// Type returns the DB backend type
func (d DatabaseConfig) Type() string {
	for k := range d {
		return k
	}
	panic("invalid database config")
}

// DSN returns the DSN
func (d DatabaseConfig) DSN() string {
	for _, v := range d {
		return v
	}
	panic("invalid database config")
}

type Duration string

func (d Duration) Duration() (time.Duration, error) {
	return time.ParseDuration(string(d))
}

// ServiceConfig represents a configuration for an HTTP server for a service
type ServiceConfig struct {
	Address        string
	ReadTimeout    Duration
	WriteTimeout   Duration
	MaxHeaderBytes int
}

// Config represents a full configuration for any paymentd related applications
type Config struct {
	// Payment config
	Payment struct {
		// Prime for obfuscating payment IDs
		PaymentIDEncPrime int64
		// XOR value to be applied to obfuscated primes
		PaymentIDEncXOR int64
	}
	// Database config
	Database struct {
		// Maximum number of retries on transaction lock errors
		TransactionMaxRetries int
		// Maximum number of database connections
		MaxOpenConns int
		// Maximum number of idle connections in the connection pool
		MaxIdleConns int
		// Principal database
		Principal struct {
			Write    DatabaseConfig
			ReadOnly DatabaseConfig
		}
		// Payment database
		Payment struct {
			Write    DatabaseConfig
			ReadOnly DatabaseConfig
		}
	}
	// API server config
	API struct {
		// Should the API server be activated?
		Active bool
		// API service config
		Service ServiceConfig

		// Should the API server provide administrative endpoints?
		ServeAdmin bool
		// SSL?
		Secure bool
		// Cookie-based authentication settings
		Cookie struct {
			// Should the API allow cookie-based authentication?
			AllowCookieAuth bool
			HTTPOnly        bool
		}

		// serve the adminpanel gui files (fullfill same origin policy)
		AdminGUIPubWWWDir string

		AuthKeys []string
	}
	// Web server config
	Web struct {
		// Whether the WWW-service should be active
		Active bool
		// The URL under which the WWW-service is served
		URL string
		// WWW service config
		Service ServiceConfig

		// Public WWW directory
		PubWWWDir string
		// Template (base-)directory
		TemplateDir string

		// Whether the WWW-service is served securely
		Secure bool

		// Cookie config
		Cookie struct {
			// Whether so serve cookies as HTTPOnly
			HTTPOnly bool
		}
		// Web auth keys for encrypting cookie auth containers
		AuthKeys []string
	}
	Provider struct {
		URL string

		ProviderTemplateDir string
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	cfg := Config{}
	cfg.Payment.PaymentIDEncPrime = 982450871
	cfg.Payment.PaymentIDEncXOR = 123456789

	cfg.Database.TransactionMaxRetries = 5
	cfg.Database.MaxOpenConns = 10
	cfg.Database.MaxIdleConns = 5

	cfg.Database.Principal.Write = NewDatabaseConfig()
	cfg.Database.Principal.Write["mysql"] = "paymentd@tcp(localhost:3306)/fritzpay_principal?charset=utf8mb4&parseTime=true&loc=UTC&timeout=1m&wait_timeout=30&interactive_timeout=30&time_zone=%22%2B00%3A00%22"

	cfg.Database.Payment.Write = NewDatabaseConfig()
	cfg.Database.Payment.Write["mysql"] = "paymentd@tcp(localhost:3306)/fritzpay_payment?charset=utf8mb4&parseTime=true&loc=UTC&timeout=1m&wait_timeout=30&interactive_timeout=30&time_zone=%22%2B00%3A00%22"

	cfg.Database.Principal.ReadOnly = nil

	cfg.API.Active = true
	cfg.API.Service.Address = ":8080"
	cfg.API.Service.ReadTimeout = Duration("10s")
	cfg.API.Service.WriteTimeout = Duration("10s")
	cfg.API.ServeAdmin = false
	cfg.API.AuthKeys = make([]string, 0)

	cfg.API.Cookie.HTTPOnly = true

	cfg.Web.URL = "http://localhost:8443"
	cfg.Web.Service.Address = ":8443"
	cfg.Web.Service.ReadTimeout = Duration("10s")
	cfg.Web.Service.WriteTimeout = Duration("10s")
	cfg.Web.AuthKeys = make([]string, 0)

	cfg.Web.Cookie.HTTPOnly = true

	cfg.Provider.URL = "http://localhost:8443"

	return cfg
}

// ReadConfig reads a JSON from the given reader into the config
func (c *Config) ReadConfig(r io.Reader) error {
	dec := json.NewDecoder(r)
	err := dec.Decode(&c)
	return err
}

// ReadConfig reads the JSON from the given reader into a new Config
func ReadConfig(r io.Reader) (Config, error) {
	cfg := &Config{}
	err := cfg.ReadConfig(r)
	return *cfg, err
}

// WriteConfig will write the given config to the given Writer as JSON (pretty printed
func WriteConfig(w io.Writer, cfg Config) error {
	jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

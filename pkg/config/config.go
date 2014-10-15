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

// ServiceConfig represents a configuration for an HTTP server for a service
type ServiceConfig struct {
	Address        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
}

// Config represents a full configuration for any paymentd related applications
type Config struct {
	// API server config
	API struct {
		// Should the API server be activated?
		Active bool
		// API service config
		Service ServiceConfig

		// Should the API server provide administrative endpoints?
		ServeAdmin bool
		// Cookie-based authentication settings
		Cookie struct {
			// Should the API allow cookie-based authentication?
			AllowCookieAuth bool
			HttpOnly        bool
			Secure          bool
		}

		AuthKeys []string
	}
	// Database config
	Database struct {
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
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	cfg := Config{}
	cfg.API.Active = true
	cfg.API.Service.Address = ":8080"
	cfg.API.Service.ReadTimeout = 10 * time.Second
	cfg.API.Service.WriteTimeout = 10 * time.Second
	cfg.API.ServeAdmin = false
	cfg.API.AuthKeys = make([]string, 0)

	cfg.API.Cookie.HttpOnly = true

	cfg.Database.Principal.Write = NewDatabaseConfig()
	cfg.Database.Principal.Write["mysql"] = "paymentd@tcp(localhost:3306)/fritzpay_principal?charset=utf8mb4&parseTime=true&loc=UTC&timeout=1m&wait_timeout=30&interactive_timeout=30"

	cfg.Database.Payment.Write = NewDatabaseConfig()
	cfg.Database.Payment.Write["mysql"] = "paymentd@tcp(localhost:3306)/fritzpay_payment?charset=utf8mb4&parseTime=true&loc=UTC&timeout=1m&wait_timeout=30&interactive_timeout=30"

	cfg.Database.Principal.ReadOnly = nil

	return cfg
}

// ReadConfig reads the JSON from the given reader into a new Config
func ReadConfig(r io.Reader) (Config, error) {
	dec := json.NewDecoder(r)
	cfg := Config{}
	err := dec.Decode(&cfg)
	return cfg, err
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

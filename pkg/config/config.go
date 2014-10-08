package config

import (
	"encoding/json"
	"io"
)

// Config represents a full configuration for any paymentd related applications
type Config struct {
	// API server config
	API struct {
		Address string
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	cfg := Config{}
	cfg.API.Address = ":8080"

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

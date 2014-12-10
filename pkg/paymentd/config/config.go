package config

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Config represents a config set
type Config map[string]string

// NewConfig creates a new config set
func NewConfig() Config {
	return Config(make(map[string]string))
}

const (
	// ConfigNameSystemPassword is the name for the system password configuration setting
	ConfigNameSystemPassword = "SystemPassword"
	// ConfigSystemPasswordBcryptCost is the cost for bcrypting the system password
	ConfigSystemPasswordBcryptCost = 10
)

const (
	// DefaultPasswordBytes is the default password length in bytes
	DefaultPasswordBytes = 32
)

type entrySetter func(Config) error

// SetPassword returns a setter for a cleartext system password
func SetPassword(pw []byte) entrySetter {
	return entrySetter(func(c Config) error {
		enc, err := bcrypt.GenerateFromPassword(pw, ConfigSystemPasswordBcryptCost)
		if err != nil {
			return err
		}
		c[ConfigNameSystemPassword] = string(enc)
		return nil
	})
}

// DefaultPassword represents a default password generation and setting
type DefaultPassword string

// Generate generates a random password
func (d *DefaultPassword) Generate() error {
	pw := make([]byte, DefaultPasswordBytes)
	_, err := rand.Read(pw)
	if err != nil {
		return err
	}

	*d = DefaultPassword(hex.EncodeToString(pw))
	return nil
}

// Entry returns a setter, which can be used with the Set() function for setting
// the default system password
func (d DefaultPassword) Entry() entrySetter {
	return SetPassword([]byte(d))
}

// DefaultPasswordSetter sets a default password
var DefaultPasswordSetter = entrySetter(func(c Config) error {
	pw := make([]byte, DefaultPasswordBytes)
	_, err := rand.Read(pw)
	if err != nil {
		return err
	}

	c[ConfigNameSystemPassword] = hex.EncodeToString(pw)
	return nil
})

// Set sets passed configuration settings
func Set(db *sql.DB, settings ...entrySetter) error {
	var err error
	// empty config
	cfg := NewConfig()
	for _, set := range settings {
		err = set(cfg)
		if err != nil {
			return err
		}
	}
	if len(cfg) == 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error on begin tx: %v", err)
	}
	err = InsertConfigTx(tx, cfg)
	if err != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			return fmt.Errorf("error on rollback tx: %v", err)
		}
		return fmt.Errorf("error on saving config: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error on commit tx: %v", err)
	}
	return nil
}

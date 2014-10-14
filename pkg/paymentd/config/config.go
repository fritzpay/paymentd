package config

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// Config represents a config set
type Config map[string]string

// NewConfig creates a new config set
func NewConfig() Config {
	return Config(make(map[string]string))
}

// InsertConfigTx saves a config set
//
// This function should be used inside a (SQL-)transaction
func InsertConfigTx(db *sql.Tx, cfg Config) error {
	stmt, err := db.Prepare(insertEntry)
	if err != nil {
		return err
	}
	t := time.Now()
	for n, v := range cfg {
		_, err = stmt.Exec(n, t, v)
		if err != nil {
			stmt.Close()
			return err
		}
	}
	stmt.Close()
	return nil
}

// InsertConfigIfNotPresentTx saves a config set if the names are not present
//
// This funtion should be used inside a (SQL-)transaction
func InsertConfigIfNotPresentTx(db *sql.Tx, cfg Config) error {
	checkExists, err := db.Prepare(selectEntryCountByName)
	if err != nil {
		return err
	}
	defer checkExists.Close()
	insert, err := db.Prepare(insertEntry)
	if err != nil {
		return err
	}
	defer insert.Close()
	var exists *sql.Row
	var numEntries int
	t := time.Now()
	for n, v := range cfg {
		exists = checkExists.QueryRow(&numEntries)
		err = exists.Scan(&numEntries)
		if err != nil {
			return err
		}
		if numEntries > 0 {
			continue
		}
		_, err = insert.Exec(n, t, v)
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	// ConfigNameSystemPassword is the name for the system password configuration setting
	ConfigNameSystemPassword = "SystemPassword"
	// ConfigSystemPasswordBcryptCost is the cost for bcrypting the system password
	ConfigSystemPasswordBcryptCost = 12
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

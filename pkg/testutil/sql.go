package testutil

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	// EnvVarMySQLTest is the environment var, which must be present to run
	// MySQL tests
	EnvVarMySQLTest = "PAYMENTD_MYSQLTEST"
	// EnvVarMySQLTestPaymentDSN holds the DSN for the test database for payment
	EnvVarMySQLTestPaymentDSN = "PAYMENTD_MYSQLTEST_PAYMENTDSN"
	// EnvVarMySQLTestPaymentDSN holds the DSN for the test database for payment
	EnvVarMySQLTestPrincipalDSN = "PAYMENTD_MYSQLTEST_PRINCIPALDSN"
)

// WithPaymentDB is a test decorator providing a DB connection to the test payment DB
func WithPaymentDB(t *testing.T, f func(db *sql.DB)) func() {
	return func() {
		if os.Getenv(EnvVarMySQLTest) == "" {
			t.Skip("Skipping MySQL test")
			return
		}
		if os.Getenv(EnvVarMySQLTestPaymentDSN) == "" {
			t.Skip("No payment DB DSN present. Skipping.")
			return
		}
		db, err := sql.Open("mysql", os.Getenv(EnvVarMySQLTestPaymentDSN))

		So(err, ShouldBeNil)
		So(db, ShouldNotBeNil)

		err = db.Ping()
		So(err, ShouldBeNil)

		f(db)
	}
}

// WithPrincipalDB is a test decorator providing a DB connection to the test principal DB
func WithPrincipalDB(t *testing.T, f func(db *sql.DB)) func() {
	return func() {
		if os.Getenv(EnvVarMySQLTest) == "" {
			t.Skip("Skipping MySQL test")
			return
		}
		if os.Getenv(EnvVarMySQLTestPrincipalDSN) == "" {
			t.Skip("No principal DB DSN present. Skipping.")
			return
		}
		db, err := sql.Open("mysql", os.Getenv(EnvVarMySQLTestPrincipalDSN))

		So(err, ShouldBeNil)
		So(db, ShouldNotBeNil)

		err = db.Ping()
		So(err, ShouldBeNil)

		f(db)
	}
}

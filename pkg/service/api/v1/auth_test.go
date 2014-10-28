package v1

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/config"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"testing"
)

func WithSystemPassword(db *sql.DB, f func()) func() {
	return func() {
		err := config.Set(db, config.SetPassword([]byte("password")))
		So(err, ShouldBeNil)

		Reset(func() {
			_, err := db.Exec(fmt.Sprintf("delete from config where name = '%s'", config.ConfigNameSystemPassword))
			So(err, ShouldBeNil)
		})

		f()
	}
}

func TestGetCredentialsWithBasicAuth(t *testing.T) {
	Convey("Given a new context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {
		ctx.Config().API.ServeAdmin = true
		So(ctx.Config().API.ServeAdmin, ShouldBeTrue)

		Convey("Given a new API service", WithService(ctx, logChan, func(s *Service, mux *http.ServeMux) {

			Convey("Given a new get credentials request", func() {
				r, err := http.NewRequest("GET", ServicePath+"/authorization", nil)
				So(err, ShouldBeNil)

				Convey("When the request method is PUT", func() {
					r.Method = "PUT"

					Convey("When the handler is called", func() {
						w := testutil.NewResponseWriter()
						mux.ServeHTTP(w, r)
						Convey("The handler should respond with method not allowed", func() {
							So(w.HeaderWritten, ShouldBeTrue)
							So(w.StatusCode, ShouldEqual, http.StatusMethodNotAllowed)
						})
					})
				})

				Convey("When the request method is DELETE", func() {
					r.Method = "DELETE"

					Convey("When the handler is called", func() {
						w := testutil.NewResponseWriter()
						mux.ServeHTTP(w, r)
						Convey("The handler should respond with method not allowed", func() {
							So(w.HeaderWritten, ShouldBeTrue)
							So(w.StatusCode, ShouldEqual, http.StatusMethodNotAllowed)
						})
					})
				})

				Convey("When the authentication method is unknown", func() {
					r.URL.Path += "/unknown"

					Convey("When the handler is called", func() {
						w := testutil.NewResponseWriter()
						mux.ServeHTTP(w, r)
						Convey("The handler should respond with a 404 (not found)", func() {
							So(w.HeaderWritten, ShouldBeTrue)
							So(w.StatusCode, ShouldEqual, http.StatusNotFound)
						})
					})
				})

				Convey("Given a payment DB", testutil.WithPaymentDB(t, func(db *sql.DB) {
					ctx.SetPaymentDB(db, nil)

					Reset(func() {
						db.Close()
					})

					Convey("Given a set system password", WithSystemPassword(db, func() {
						Convey("When retrieving a basic authorization", func() {
							r.Method = "GET"
							r.URL.Path += "/basic"

							Convey("When using the correct password", func() {
								r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("root:password")))

								Convey("When the handler is called", func() {
									w := testutil.NewResponseWriter()
									mux.ServeHTTP(w, r)
									Convey("The handler should respond with OK", func() {
										So(w.HeaderWritten, ShouldBeTrue)
										So(w.StatusCode, ShouldEqual, http.StatusOK)
										Convey("The body should contain the authorization container", func() {
											m := make(map[string]string)
											dec := json.NewDecoder(&w.Buf)
											err := dec.Decode(&m)
											So(err, ShouldBeNil)
											So(m["Authorization"], ShouldNotBeEmpty)
										})
									})

									Convey("Given the returned authorization container", func() {
										m := make(map[string]string)
										dec := json.NewDecoder(&w.Buf)
										err := dec.Decode(&m)
										So(err, ShouldBeNil)
										So(m["Authorization"], ShouldNotBeEmpty)

										Convey("Given a service request context", func() {
											service.SetRequestContext(r, ctx)

											Convey("Given a get user request", func() {
												r.Method = "GET"
												r.URL.Path = ServicePath + "/user"
												r.Header.Set("Authorization", m["Authorization"])

												Convey("When the handler is called", func() {
													w := testutil.NewResponseWriter()
													mux.ServeHTTP(w, r)
													Convey("The handler should respond with OK", func() {
														So(w.HeaderWritten, ShouldBeTrue)
														So(w.StatusCode, ShouldEqual, http.StatusOK)
													})
												})
											})
										})
									})
								})
							})

							Convey("When using a wrong password", func() {
								r.Header.Set("Authorization", "Basic dede")

								Convey("When the handler is called", func() {
									w := testutil.NewResponseWriter()
									mux.ServeHTTP(w, r)
									Convey("The handler should respond with Unauthorized", func() {
										So(w.HeaderWritten, ShouldBeTrue)
										So(w.StatusCode, ShouldEqual, http.StatusUnauthorized)
									})
								})
							})

							Convey("When using a bad authorization header", func() {
								r.Header.Set("Authorization", "Basic")

								Convey("When the handler is called", func() {
									w := testutil.NewResponseWriter()
									mux.ServeHTTP(w, r)
									Convey("The handler should request an authorization", func() {
										So(w.HeaderWritten, ShouldBeTrue)
										So(w.StatusCode, ShouldEqual, http.StatusUnauthorized)
										So(w.Header().Get("WWW-Authenticate"), ShouldNotBeEmpty)
									})
								})
							})
						})
					}))
				}))
			})
		}))
	}))
}

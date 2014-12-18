package v1

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/principal"

	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/testutil"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
)

// WithService is a test decorator and will provide a service instance and a mux router
// for the given service context
func WithService(ctx *service.Context, logChan <-chan *log15.Record, f func(s *Service, mux *mux.Router)) func() {
	return func() {
		testMsg := "testmsg"
		ctx.Log().Info(testMsg)
		logMsg := <-logChan

		So(logMsg.Msg, ShouldEqual, testMsg)

		mux := mux.NewRouter()
		service, err := NewService(ctx, mux)
		So(err, ShouldBeNil)

		f(service, mux)
	}
}

func WithAuthorization(mx *mux.Router, f func(auth string)) func() {
	return func() {
		req, err := http.NewRequest("GET", ServicePath+"/authorization/basic", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("root:password")))

		w := testutil.NewResponseWriter()
		mx.ServeHTTP(w, req)

		So(w.HeaderWritten, ShouldBeTrue)
		So(w.StatusCode, ShouldEqual, http.StatusOK)

		m := make(map[string]string)
		dec := json.NewDecoder(&w.Buf)
		err = dec.Decode(&m)
		So(err, ShouldBeNil)

		f(m["Authorization"])
	}
}

func TestServiceSetup(t *testing.T) {
	Convey("Given a new context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {

		Convey("When the admin API is active", func() {
			ctx.Config().API.ServeAdmin = true

			Convey("Given a new service", WithService(ctx, logChan, func(s *Service, mx *mux.Router) {

				Convey("The admin API routes should be registered", func() {
					r, err := http.NewRequest("GET", ServicePath+"/authorization", nil)
					So(err, ShouldBeNil)

					rm := mux.RouteMatch{}
					match := mx.Match(r, &rm)

					So(match, ShouldBeTrue)
				})
			}))
		})

		Convey("When the config does not request the admin API to be active", func() {
			ctx.Config().API.ServeAdmin = false

			So(ctx.Config().API.ServeAdmin, ShouldBeFalse)

			Convey("Given a new service", WithService(ctx, logChan, func(s *Service, mx *mux.Router) {

				Convey("The admin registered log message should not be present", func() {
					var logMessagePresent bool
				drain:
					for {
						select {
						case msg := <-logChan:
							if strings.Contains(msg.Msg, "admin API") {
								logMessagePresent = true
							}
						default:
							So(logMessagePresent, ShouldBeFalse)
							break drain
						}
					}
				})

				Convey("Then the admin API routes should not be registered", func() {
					r, err := http.NewRequest("GET", ServicePath+"/authorization", nil)
					So(err, ShouldBeNil)

					rm := mux.RouteMatch{}
					match := mx.Match(r, &rm)

					So(match, ShouldBeFalse)
				})
			}))
		})
	}))
}

func TestGetCredentialsWithBasicAuth(t *testing.T) {
	Convey("Given a new context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {
		ctx.Config().API.ServeAdmin = true
		So(ctx.Config().API.ServeAdmin, ShouldBeTrue)

		Convey("Given a new API service", WithService(ctx, logChan, func(s *Service, mx *mux.Router) {

			Convey("Given a new get credentials request", func() {
				r, err := http.NewRequest("GET", ServicePath+"/authorization", nil)
				So(err, ShouldBeNil)

				Convey("When the request method is POST", func() {
					r.Method = "POST"

					Convey("When the handler is called", func() {
						w := testutil.NewResponseWriter()
						mx.ServeHTTP(w, r)
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
						mx.ServeHTTP(w, r)
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

					Convey("When retrieving a basic authorization", func() {
						r.Method = "GET"
						r.URL.Path += "/basic"

						Convey("When using the correct password", func() {
							r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("root:password")))

							Convey("When the handler is called", func() {
								w := testutil.NewResponseWriter()
								mx.ServeHTTP(w, r)
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

								Convey("Given the returned (correct) authorization container", func() {
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
												mx.ServeHTTP(w, r)
												Convey("The handler should respond with OK", func() {
													So(w.HeaderWritten, ShouldBeTrue)
													So(w.StatusCode, ShouldEqual, http.StatusOK)
												})
											})
										})
									})
								})
							})

							Convey("Given cookie auth is allowed", func() {
								ctx.Config().API.Cookie.AllowCookieAuth = true

								Convey("When the handler is called", func() {
									w := testutil.NewResponseWriter()
									mx.ServeHTTP(w, r)
									Convey("The handler should set a cookie", func() {
										So(w.Header().Get("Set-Cookie"), ShouldNotBeEmpty)
									})
								})
							})
						})

						Convey("When using a wrong password", func() {
							r.Header.Set("Authorization", "Basic dede")

							Convey("When the handler is called", func() {
								w := testutil.NewResponseWriter()
								mx.ServeHTTP(w, r)
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
								mx.ServeHTTP(w, r)
								Convey("The handler should request an authorization", func() {
									So(w.HeaderWritten, ShouldBeTrue)
									So(w.StatusCode, ShouldEqual, http.StatusUnauthorized)
									So(w.Header().Get("WWW-Authenticate"), ShouldNotBeEmpty)
								})
							})
						})
					})
				}))
			})
		}))
	}))
}

func TestInitPaymentRequest(t *testing.T) {
	Convey("Given a init payment request", t, func() {
		req := InitPaymentRequest{}

		Convey("When populated with test values", func() {
			req.ProjectKey = "abcdef123456"
			req.Ident = "testIdent"
			req.Amount.Int64 = 1234
			req.Subunits.Int8 = 2
			req.Currency = "EUR"
			req.Country = "DE"
			req.PaymentMethodID = 12
			req.Timestamp = 1234567
			req.Nonce = "testNonce"

			Convey("When creating a signature base string", func() {
				sig, err := req.Message()
				So(err, ShouldBeNil)

				Convey("It should match the expected signature", func() {
					expected := "abcdef123456testIdent12342EURDE121234567testNonce"
					So(string(sig), ShouldEqual, expected)
				})
			})

			Convey("When adding the optional locale value", func() {
				req.Locale = "en_US"

				Convey("When creating a signature base string", func() {
					sig, err := req.Message()
					So(err, ShouldBeNil)

					Convey("It should match the expected signature", func() {
						expected := "abcdef123456testIdent12342EURDE12en_US1234567testNonce"
						So(string(sig), ShouldEqual, expected)
					})
				})
			})
		})
	})
}

func TestInitPaymentRequestValidation(t *testing.T) {
	Convey("Given a init payment request", t, func() {
		req := InitPaymentRequest{}

		Convey("When populated with test values", func() {
			req.ProjectKey = "abcdef123456"
			req.Ident = "testIdent"
			req.Amount.Int64, req.Amount.Set = 1234, true
			req.Subunits.Int8, req.Subunits.Set = 2, true
			req.Currency = "EUR"
			req.Country = "DE"
			req.PaymentMethodID = 12
			req.Timestamp = 1234567
			req.Nonce = "testNonce"

			Convey("When validating without project key", func() {
				req.ProjectKey = ""
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "ProjectKey")
				})
			})

			Convey("When validating without ident", func() {
				req.Ident = ""
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Ident")
				})
			})

			Convey("When validating with a too large ident", func() {
				req.Ident = strings.Repeat("s", 200)
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Ident")
				})
			})

			Convey("When validating without an Amount", func() {
				req.Amount.Set = false
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Amount")
				})
			})

			Convey("When validating with a negative Amount", func() {
				req.Amount.Int64 = -1000
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Amount")
				})
			})

			Convey("When validating without a Subunit", func() {
				req.Subunits.Set = false
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Subunits")
				})
			})
		})
	})
}

func TestGetProvider(t *testing.T) {
	Convey("Given a test context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {
		ctx.Config().API.ServeAdmin = true

		Convey("Given a service", WithService(ctx, logChan, func(s *Service, mx *mux.Router) {

			Convey("Given a request for the test provider", func() {
				req, err := http.NewRequest("GET", ServicePath+"/provider/fritzpay", nil)
				So(err, ShouldBeNil)

				rm := mux.RouteMatch{}
				match := mx.Match(req, &rm)
				So(match, ShouldBeTrue)

				Convey("Given a payment db", testutil.WithPaymentDB(t, func(db *sql.DB) {
					ctx.SetPaymentDB(db, nil)
					Reset(func() { db.Close() })

					Convey("Given a valid authorization", WithAuthorization(mx, func(auth string) {
						req.Header.Set("Authorization", auth)

						Convey("When executing the request", func() {
							w := testutil.NewResponseWriter()
							mx.ServeHTTP(w, req)

							Convey("It should succeed", func() {
								So(w.HeaderWritten, ShouldBeTrue)
								So(w.StatusCode, ShouldEqual, http.StatusOK)
							})
						})
					}))
				}))
			})
		}))
	}))
}

func WithCreatePrincipalRequest(ctx *service.Context, pr principal.Principal, f func(req *http.Request)) func() {
	return func() {
		jsonB, err := json.Marshal(pr)
		So(err, ShouldBeNil)
		buf := bytes.NewBuffer(jsonB)

		req, err := http.NewRequest("PUT", ServicePath+"/principal", buf)
		So(err, ShouldBeNil)

		// request context
		service.SetRequestContext(req, ctx)
		Reset(func() {
			service.ClearRequestContext(req)
		})

		f(req)
	}
}

func TestPrincipal(t *testing.T) {
	Convey("Given a test context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {
		ctx.Config().API.ServeAdmin = true

		Convey("Given a service", WithService(ctx, logChan, func(s *Service, mx *mux.Router) {

			Convey("Given a payment db", testutil.WithPaymentDB(t, func(db *sql.DB) {
				ctx.SetPaymentDB(db, nil)
				Reset(func() { db.Close() })

				Convey("Given a principal db", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {
					ctx.SetPrincipalDB(prDB, nil)
					Reset(func() { prDB.Close() })

					Convey("Given a principal", func() {
						pr := principal.Principal{
							Name: fmt.Sprintf("test%d", time.Now().UnixNano()),
							Metadata: map[string]string{
								"test": "one",
							},
						}

						Convey("Given the principal has a valid status", func() {
							pr.Status = principal.PrincipalStatusActive

							Convey("Given a create principal request", WithCreatePrincipalRequest(ctx, pr, func(req *http.Request) {
								rm := mux.RouteMatch{}
								match := mx.Match(req, &rm)
								So(match, ShouldBeTrue)

								Convey("Given a valid authorization", WithAuthorization(mx, func(auth string) {
									req.Header.Set("Authorization", auth)

									Convey("When executing the request", func() {
										w := testutil.NewResponseWriter()
										mx.ServeHTTP(w, req)

										Convey("It should succeed", func() {
											So(w.HeaderWritten, ShouldBeTrue)
											So(w.StatusCode, ShouldEqual, http.StatusOK)

											resp := ServiceResponse{}
											dec := json.NewDecoder(&w.Buf)
											err := dec.Decode(&resp)
											So(err, ShouldBeNil)
											So(resp.Status, ShouldEqual, StatusSuccess)

											respPr, ok := resp.Response.(map[string]interface{})
											So(ok, ShouldBeTrue)

											principalIDStr := respPr["ID"].(string)
											principalID, err := strconv.ParseInt(principalIDStr, 10, 64)
											So(err, ShouldBeNil)
											So(principalID, ShouldNotEqual, 0)

											Convey("Given an update request", func() {
												pr.Metadata["test2"] = "two"
												jsonB, err := json.Marshal(pr)
												So(err, ShouldBeNil)
												buf := bytes.NewBuffer(jsonB)
												req, err = http.NewRequest("POST", ServicePath+"/principal/"+pr.Name, buf)
												So(err, ShouldBeNil)
												req.Header.Set("Authorization", auth)

												rm := mux.RouteMatch{}
												match := mx.Match(req, &rm)
												So(match, ShouldBeTrue)

												service.SetRequestContext(req, ctx)
												Reset(func() {
													service.ClearRequestContext(req)
												})

												Convey("When executing the request", func() {
													w := testutil.NewResponseWriter()
													mx.ServeHTTP(w, req)

													Convey("It should succeed", func() {
														So(w.HeaderWritten, ShouldBeTrue)
														So(w.StatusCode, ShouldEqual, http.StatusOK)

														resp := ServiceResponse{}
														dec := json.NewDecoder(&w.Buf)
														err := dec.Decode(&resp)
														So(err, ShouldBeNil)
														So(resp.Status, ShouldEqual, StatusSuccess)

														respPr, ok := resp.Response.(map[string]interface{})
														So(ok, ShouldBeTrue)
														So(respPr["Metadata"], ShouldNotBeNil)

														meta, ok := respPr["Metadata"].(map[string]interface{})
														So(ok, ShouldBeTrue)

														So(meta["test2"], ShouldEqual, "two")
													})
												})
											})
										})
									})
								}))
							}))
						})

						Convey("Given the principal has an invalid status", func() {
							pr.Status = "invalidfwefwf"
							Convey("Given a create principal request", WithCreatePrincipalRequest(ctx, pr, func(req *http.Request) {
								rm := mux.RouteMatch{}
								match := mx.Match(req, &rm)
								So(match, ShouldBeTrue)

								Convey("Given a valid authorization", WithAuthorization(mx, func(auth string) {
									req.Header.Set("Authorization", auth)

									Convey("When executing the request", func() {
										w := testutil.NewResponseWriter()
										mx.ServeHTTP(w, req)

										Convey("It should return a bad request", func() {
											So(w.HeaderWritten, ShouldBeTrue)
											So(w.StatusCode, ShouldEqual, http.StatusBadRequest)
										})
									})
								}))
							}))
						})
					})
				}))
			}))
		}))
	}))
}

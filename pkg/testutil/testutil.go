package testutil

import (
	"bytes"
	"net/http"

	"github.com/fritzpay/paymentd/pkg/config"
	"github.com/fritzpay/paymentd/pkg/service"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	logChanBufferSize = 32
)

// WithContext is a decorator for GoConvey based tests
//
// It will inject a service context and a log channel, where log messages can be read from
func WithContext(f func(*service.Context, <-chan *log15.Record)) func() {
	return func() {
		logChan := make(chan *log15.Record, logChanBufferSize)
		log := log15.New()
		log.SetHandler(log15.ChannelHandler(logChan))

		var cancel context.CancelFunc
		baseCtx := context.Background()
		baseCtx, cancel = context.WithCancel(baseCtx)

		ctx, err := service.NewContext(context.Background(), config.DefaultConfig(), log)
		So(err, ShouldBeNil)

		ctx.APIKeychain().AddBinKey([]byte("test"))
		So(ctx.APIKeychain().KeyCount(), ShouldEqual, 1)

		f(ctx, logChan)

		Reset(func() {
			cancel()
		})
	}
}

// ResponseWriter is a mock http.ResponseWriter which can be used to inspect
// HTTP handler responses
type ResponseWriter struct {
	Buf           bytes.Buffer
	H             http.Header
	HeaderWritten bool
	StatusCode    int
}

// NewResponseWriter creates a response writer to capture http handler responses
func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		H:          http.Header(make(map[string][]string)),
		StatusCode: http.StatusOK,
	}
}

// Header implementing the http.ResponseWriter
func (r *ResponseWriter) Header() http.Header {
	return r.H
}

// Write implementing the http.ResponseWriter, io.Writer
func (r *ResponseWriter) Write(p []byte) (int, error) {
	if !r.HeaderWritten {
		r.HeaderWritten = true
	}
	return (&(r.Buf)).Write(p)
}

// WriteHeader implementing the http.ResponseWriter
func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.HeaderWritten = true
	r.StatusCode = statusCode
}

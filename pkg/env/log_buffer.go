// +build go1.3

package env

import (
	"bytes"
	"sync"

	"gopkg.in/inconshreveable/log15.v2"
)

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, logBufferSize))
	},
}

// DaemonFormat returns a log15.Format, which produces records which can be forwarded to
// syslog by the init system
func DaemonFormat() log15.Format {
	return log15.FormatFunc(func(r *log15.Record) []byte {
		common := []interface{}{r.KeyNames.Time, r.Time, r.KeyNames.Lvl, r.Lvl, r.KeyNames.Msg, r.Msg}
		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		logLevel(buf, r.Lvl)
		logRecord(buf, append(common, r.Ctx...))
		b := buf.Bytes()
		bufferPool.Put(buf)
		return b
	})
}

func escapeString(s string) string {
	needQuotes := false
	e := bufferPool.Get().(*bytes.Buffer)
	e.Reset()
	e.WriteByte('"')
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' {
			needQuotes = true
		}

		switch r {
		case '\\', '"':
			e.WriteByte('\\')
			e.WriteByte(byte(r))
		case '\n':
			e.WriteByte('\\')
			e.WriteByte('n')
		case '\r':
			e.WriteByte('\\')
			e.WriteByte('r')
		case '\t':
			e.WriteByte('\\')
			e.WriteByte('t')
		default:
			e.WriteRune(r)
		}
	}
	e.WriteByte('"')
	start, stop := 0, e.Len()
	if !needQuotes {
		start, stop = 1, stop-1
	}
	eStr := string(e.Bytes()[start:stop])
	bufferPool.Put(e)
	return eStr
}

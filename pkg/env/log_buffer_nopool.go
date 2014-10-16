// +build go1.1,!go1.3 go1.2,!go1.3

package env

import (
	"bytes"
	"gopkg.in/inconshreveable/log15.v2"
	"sync"
)

// DaemonFormat returns a log15.Format, which produces records which can be forwarded to
// syslog by the init system
func DaemonFormat() log15.Format {
	return log15.FormatFunc(func(r *log15.Record) (b []byte) {
		common := []interface{}{r.KeyNames.Time, r.Time, r.KeyNames.Lvl, r.Lvl, r.KeyNames.Msg, r.Msg}
		buf := bytes.NewBuffer(nil)
		logLevel(buf, r.Lvl)
		logRecord(buf, append(common, r.Ctx...))
		b = buf.Bytes()
		return
	})
}

func escapeString(s string) (eStr string) {
	needQuotes := false
	e := bytes.NewBuffer(nil)
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
	eStr = string(e.Bytes()[start:stop])
	return
}

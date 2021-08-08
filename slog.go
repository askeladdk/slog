// Package slog implements structured logging for lazy people.
package slog

import (
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

const (
	// Lcolor enables colorized output to the terminal.
	Lcolor = 1 << (iota + 16)
	// Lparsefields enables parsing the message for key-value fields.
	Lparsefields
	// LstdFlags defines an initial set of flags.
	LstdFlags = log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lmsgprefix | Lcolor | Lparsefields
)

func scanKeyVals(s string) (i int, key, val string, quote, ok bool) {
	var start, width int
	// Skip leading spaces
	for ; i < len(s); i += width {
		var r rune
		r, width = utf8.DecodeRuneInString(s[i:])
		if !unicode.IsSpace(r) {
			break
		}
	}
	// Scan until equals without space, marking end of key
	for start = i; i < len(s); i += width {
		var r rune
		r, width = utf8.DecodeRuneInString(s[i:])
		if unicode.IsSpace(r) {
			return
		} else if r == '=' {
			i, key = i+width, s[start:i]
			break
		}
	}
	// Check for eol
	if i == len(s) {
		return
	}
	// Scan until quote, marking end of quoted value
	if s[i] == '"' {
		i++
		for start = i; i < len(s); i += width {
			var r rune
			r, width = utf8.DecodeRuneInString(s[i:])
			if r == '"' {
				i, val, quote, ok = i+width, s[start:i], true, true
				return
			}
		}
		return
	}
	// Scan until space, marking end of value
	for start = i; i < len(s); i += width {
		var r rune
		r, width = utf8.DecodeRuneInString(s[i:])
		if unicode.IsSpace(r) {
			i, val, ok = i+width, s[start:i], true
			return
		}
	}
	// End of line, marking end of value
	i, val, ok = len(s), s[start:], true
	return
}

type colorFunc func([]byte, string) []byte

func color(dst []byte, c string) []byte { return append(dst, c...) }
func plain(dst []byte, _ string) []byte { return dst }

const keycol = "\033[34m"
const strcol = "\033[32m"
const clrcol = "\033[0m"

func appendKey(dst []byte, s string, col colorFunc) []byte {
	dst = col(dst, keycol)
	dst = strconv.AppendQuote(dst, s)
	dst = col(dst, clrcol)
	dst = append(dst, ':')
	return dst
}

func appendVal(dst []byte, s string, col colorFunc) []byte {
	dst = append(dst, s...)
	return dst
}

func appendQuote(dst []byte, s string, col colorFunc) []byte {
	dst = col(dst, strcol)
	dst = strconv.AppendQuote(dst, s)
	dst = col(dst, clrcol)
	return dst
}

func appendFloat(dst []byte, flt float64, col colorFunc) []byte {
	dst = strconv.AppendFloat(dst, flt, 'f', -1, 64)
	return dst
}

func appendInt(dst []byte, i int64, col colorFunc) []byte {
	dst = strconv.AppendInt(dst, i, 10)
	return dst
}

func appendKeyVal(dst []byte, col colorFunc, key, val string, quote bool) []byte {
	dst = append(dst, ',')
	dst = appendKey(dst, key, col)

	// string
	if quote {
		return appendQuote(dst, val, col)
	}

	// keyword
	switch val {
	case "true", "false", "null":
		return appendVal(dst, val, col)
	case "nil":
		return appendVal(dst, "null", col)
	}

	// number
	if strings.ContainsAny(val, "0123456789") {
		if strings.IndexByte(val, '.') >= 0 {
			if flt, err := strconv.ParseFloat(val, 64); err == nil {
				return appendFloat(dst, flt, col)
			}
		} else {
			if i, err := strconv.ParseInt(val, 0, 64); err == nil {
				return appendInt(dst, i, col)
			}
		}
	}

	// string
	return appendQuote(dst, val, col)
}

func parselog(dst []byte, col colorFunc, text, prefix string, flags int) []byte {
	dst = append(dst, '{')

	text = strings.TrimRightFunc(text, unicode.IsSpace)

	// prefix
	if prefix != "" && flags&log.Lmsgprefix == 0 {
		text = text[len(prefix):]
		prefix = strings.Trim(prefix, "\t ,.:;[]")
		if prefix != "" {
			dst = appendKey(dst, "prfx", col)
			dst = appendQuote(dst, prefix, col)
			dst = append(dst, ',')
		}
	}

	// date and time
	if flags&(log.Ldate|log.Ltime|log.Lmicroseconds) != 0 {
		dst = appendKey(dst, "time", col)
		dst = col(dst, strcol)
		dst = append(dst, '"')
		if flags&log.Ldate != 0 {
			ofs := len(dst)
			dst, text = append(dst, text[:11]...), text[11:]
			dst[ofs+4] = '-'
			dst[ofs+7] = '-'
			dst[ofs+10] = 'T'
		}
		if flags&log.Ltime != 0 {
			n := 8
			if flags&log.Lmicroseconds != 0 {
				n += 7
			}
			dst, text = append(dst, text[:n]...), text[n+1:]
		}
		if flags&(log.Ldate|log.Ltime|log.LUTC) == log.Ldate|log.Ltime|log.LUTC {
			dst = append(dst, 'Z')
		}
		dst = append(dst, '"')
		dst = col(dst, clrcol)
		dst = append(dst, ',')
	}

	// file name and line number
	if flags&(log.Llongfile|log.Lshortfile) != 0 {
		var file, line string
		i := strings.IndexByte(text, ':')
		file, text = text[:i], text[i+1:]
		dst = appendKey(dst, "fnam", col)
		dst = appendQuote(dst, file, col)
		dst = append(dst, ',')
		dst = appendKey(dst, "flno", col)
		i = strings.IndexByte(text, ':')
		line, text = text[:i], text[i+2:]
		dst = appendVal(dst, line, col)
		dst = append(dst, ',')
	}

	// message
	dst = appendKey(dst, "mesg", col)
	dst = appendQuote(dst, text, col)

	// fields
	if flags&Lparsefields != 0 && strings.IndexByte(text, '=') >= 0 {
		for len(text) > 0 {
			i, key, val, quote, ok := scanKeyVals(text)
			text = text[i:]
			if ok {
				dst = appendKeyVal(dst, col, key, val, quote)
			}
		}
	}

	return append(dst, "}\n"...)
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

func zcstring(p []byte) string { return *(*string)(unsafe.Pointer(&p)) }

func isterm(w io.Writer) (term bool) {
	if f, ok := w.(interface{ Stat() (os.FileInfo, error) }); ok {
		stat, _ := f.Stat()
		term = stat != nil && stat.Mode()&os.ModeCharDevice != 0
	}
	return
}

// NewWriter creates a new structured logging output writer.
// The prefix and flags of the logger must not be changed afterwards.
func NewWriter(l *log.Logger, w io.Writer) io.Writer {
	if w == io.Discard {
		return io.Discard
	}

	prefix, flags := l.Prefix(), l.Flags()
	pbuf := make([]byte, 0, 256)
	col := plain

	if l.Flags()&Lcolor != 0 && isterm(w) {
		col = color
	}

	return writerFunc(func(p []byte) (int, error) {
		pbuf = parselog(pbuf[:0], col, zcstring(p), prefix, flags)
		if _, err := w.Write(pbuf); err != nil {
			return 0, err
		}
		return len(p), nil
	})
}

// New creates a new log.Logger that produces structured logs.
// The prefix and flags of the logger must not be changed afterwards.
func New(w io.Writer, prefix string, flag int) *log.Logger {
	l := log.New(nil, prefix, flag)
	l.SetOutput(NewWriter(l, w))
	return l
}

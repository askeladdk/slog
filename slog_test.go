package slog

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"
	"time"
)

func TestScanKeyVals(t *testing.T) {
	if _, k, v, q, ok := scanKeyVals("hello=world"); k != "hello" || v != "world" || q || !ok {
		t.Fatal()
	}

	if _, k, v, q, ok := scanKeyVals("hello=world "); k != "hello" || v != "world" || q || !ok {
		t.Fatal()
	}

	if _, _, _, _, ok := scanKeyVals("hello="); ok {
		t.Fatal()
	}

	if _, _, _, _, ok := scanKeyVals("hello = world"); ok {
		t.Fatal()
	}

	if _, k, v, q, ok := scanKeyVals("h=\"ello world\""); k != "h" || v != "ello world" || !q || !ok {
		t.Fatal()
	}

	if _, _, _, _, ok := scanKeyVals("h=\"ello world"); ok {
		t.Fatal()
	}
}

func TestParse(t *testing.T) {
	var struc struct {
		Prefix  string    `json:"prfx"`
		Time    time.Time `json:"time"`
		File    string    `json:"fnam"`
		Line    int       `json:"flno"`
		Message string    `json:"mesg"`
		A       string    `json:"a"`
		B       int       `json:"b"`
		C       bool      `json:"c"`
		D       float64   `json:"d"`
		E       string    `json:"e"`
		F       string    `json:"f"`
	}

	var b bytes.Buffer
	mesg := "a=\"hello world\" b=1337 c=true d=3.14 e=/index.html f=nil"
	l := New(&b, "test: ", log.Ldate|log.Ltime|log.LUTC|log.Lmicroseconds|log.Lshortfile|Lparsefields)
	l.Println(mesg)

	res := b.Bytes()

	if err := json.Unmarshal(res, &struc); err != nil {
		t.Fatal(err)
	}

	if struc.Prefix != "test" {
		t.Fatal()
	} else if struc.Time.IsZero() {
		t.Fatal()
	} else if struc.File != "slog_test.go" {
		t.Fatal()
	} else if struc.Line == 0 {
		t.Fatal()
	} else if struc.Message != mesg {
		t.Fatal()
	} else if struc.A != "hello world" {
		t.Fatal()
	} else if struc.B != 1337 {
		t.Fatal()
	} else if struc.C != true {
		t.Fatal()
	} else if struc.D != 3.14 {
		t.Fatal()
	} else if struc.E != "/index.html" {
		t.Fatal()
	} else if struc.F != "" {
		t.Fatal()
	}
}

func BenchmarkStdLogger(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 2<<20))
	l := log.New(buf, "test: ", log.Ldate|log.Ltime|log.LUTC|log.Lmicroseconds|log.Lshortfile)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Println("a=\"hello world\" b=1337 c=true d=3.14 e=/index.html f=nil")
	}
	b.StopTimer()
}

func BenchmarkSlog(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 2<<20))
	l := New(buf, "test: ", log.Ldate|log.Ltime|log.LUTC|log.Lmicroseconds|log.Lshortfile)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Println("a=\"hello world\" b=1337 c=true d=3.14 e=/index.html f=nil")
	}
	b.StopTimer()
}

func BenchmarkSlogParseFields(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 2<<20))
	l := New(buf, "test: ", log.Ldate|log.Ltime|log.LUTC|log.Lmicroseconds|log.Lshortfile|Lparsefields)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Println("a=\"hello world\" b=1337 c=true d=3.14 e=/index.html f=nil")
	}
	b.StopTimer()
}

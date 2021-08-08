# slog - structured logging for lazy gophers

[![GoDoc](https://godoc.org/github.com/askeladdk/slog?status.png)](https://godoc.org/github.com/askeladdk/slog)
[![Go Report Card](https://goreportcard.com/badge/github.com/askeladdk/slog)](https://goreportcard.com/report/github.com/askeladdk/slog)

## Overview

Slog parses and converts log messages produced by the standard logger to JSON objects. Any `key=value` text fragments found in the message are extracted as separate JSON fields. No boilerplate, just `Printf`.

## Install

```
go get -u github.com/askeladdk/slog
```

## Quickstart

Use the function `slog.New` to create a `log.Logger` that produces structured logs. It has the same signature as `log.New` in the standard library and is backwards compatible.

Enable all features and create a logger:

```go
logger := slog.New(os.StdErr, "level=info ", slog.LstdFlags)
```

Log an event:

```go
logger.Printf("requested url=%s with method=%s with response status=%d", "/index.html", "GET", 200)
```

Result:

```json
{"time":"2021-08-08T19:06:35.252044Z","mesg":"level=info requested url=/index.html with method=GET with response status=200","level":"info","url":"/index.html","method":"GET","status":200}
```

Use `slog.NewWriter` to create a new structured writer and attach it to the default logger with `SetOutput`:

```go
log.SetFlags(slog.LstdFlags)
log.SetOutput(slog.NewWriter(os.StdErr, log.Default()))
log.Println("hello world")
```

Note that the logger flags and prefix must not be changed after a writer has been created.

Read the rest of the [documentation on pkg.go.dev](https://pkg.go.dev/github.com/askeladdk/slog). It's easy-peasy!

## Performance

Unscientific benchmarks on my laptop suggest that slog is about twice
as memory and thrice as CPU intensive as the standard logger by itself.

```
% go test -bench=. -benchmem -benchtime=1000000x
goos: darwin
goarch: amd64
pkg: github.com/askeladdk/slog
cpu: Intel(R) Core(TM) i5-5287U CPU @ 2.90GHz
BenchmarkStdLogger-4          	 1000000	      1555 ns/op	     544 B/op	       3 allocs/op
BenchmarkSlog-4              	 1000000	      3251 ns/op	     860 B/op	       5 allocs/op
BenchmarkSlogParseFields-4   	 1000000	      4556 ns/op	     860 B/op	       5 allocs/op
```

## License

Package slog is released under the terms of the ISC license.

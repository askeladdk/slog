# slog - structured logging for lazy gophers

[![GoDoc](https://godoc.org/github.com/askeladdk/slog?status.png)](https://godoc.org/github.com/askeladdk/slog)
[![Go Report Card](https://goreportcard.com/badge/github.com/askeladdk/slog)](https://goreportcard.com/report/github.com/askeladdk/slog)

## Overview

Want to have structured logging but don't want to deal with all the formalities that come with it? Slog has got you covered.

Slog is a drop-in adapter for turning the standard logger `log.Logger` into a structured logger that produces JSON objects. It requires almost no configuration and extracts `key=value` fragments embedded in log messages as separate JSON fields. Write log messages in a natural way and let slog take care of structuring it for you.

## Install

```
go get -u github.com/askeladdk/slog
```

## Quickstart

Use the function `slog.New` to create a new structured logger.
It has the same signature as `log.New` in the standard library.

Create logger and enable all features:

```go
logger := slog.New(os.StdErr, "level=INFO ", slog.LstdFlags)
```

Log an event:

```go
logger.Printf("requested url=%s with method=%s with response status=%d", "/index.html", "GET", 200)
```

Result:

```json
{"time":"2021-08-08T19:06:35.252044Z","mesg":"level=info requested url=/index.html with method=GET with response status=200","level":"info","url":"/index.html","method":"GET","status":200}
```

Use `slog.NewWriter` to create a new writer and attach it to a logger with `SetOutput`:

```go
log.SetFlags(slog.LstdFlags)
log.SetOutput(slog.NewWriter(log.Default(), os.StdErr))
log.Println("hello world")
```

Note that the logger flags and prefix must not be changed after a writer has been created.

## Configuration

Like the standard logger, slog is configured via flags. It uses all the standard flags and introduces two new ones, `slog.Lcolor` and `slog.Lparsefields`.

Flag `slog.Lcolor` colorizes the output if the output writer is detected to be a tty.

Flag `slog.Lparsefields` parses the log message (including prefix if `log.Lmsgprefix` is set) for key-value pairs and stores them as separate fields in the JSON object. A key-value pair is any fragment of text of the form `key=value` or `key="another value"`. The key cannot contain spaces and the equals sign cannot be surrounded by spaces. Slog does not check if there are duplicate field names.

The standard logger produces non-standard timestamps.
Slog converts the timestamps to RFC3339 format if the
flags `log.Ldate`, `log.Ltime` and `log.LUTC` are all set
and stores it in the `time` field.

The prefix, if any, is parsed differently depending on whether `log.Lmsgprefix` is set.
If it is not, then the prefix is trimmed of spaces and punctuation marks and stored in the `prfx` field.
If it is, then it is considered part of the log message. The log message is stored in the `mesg` field.

If flags `log.Llongfile` or `log.Lshortfile` as set, slog parses the file name and line number in two separate fields `fnam` and `flno`.

Read the rest of the [documentation on pkg.go.dev](https://pkg.go.dev/github.com/askeladdk/slog). It's easy-peasy!

## Performance

Unscientific benchmarks on my laptop suggest that slog is about twice
as memory and thrice as CPU intensive as the standard logger by itself,
but it is still very low in absolute terms.

```
% go test -bench=. -benchmem -benchtime=1000000x
goos: darwin
goarch: amd64
pkg: github.com/askeladdk/slog
cpu: Intel(R) Core(TM) i5-5287U CPU @ 2.90GHz
BenchmarkBaseLine-4          	 1000000	      1555 ns/op	     544 B/op	       3 allocs/op
BenchmarkSlog-4              	 1000000	      3251 ns/op	     860 B/op	       5 allocs/op
BenchmarkSlogParseFields-4   	 1000000	      4556 ns/op	     860 B/op	       5 allocs/op
```

## License

Package slog is released under the terms of the ISC license.

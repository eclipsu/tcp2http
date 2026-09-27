# HTTP from TCP

I got tired of treating HTTP as a library call. I could send requests all day and still not really know what one looks like on the wire.

This repo is how I am learning that. I read bytes from a TCP socket, find the request line, split headers on `\r\n`, and read a body only because `Content-Length` told me how long it is. I write the response the same way, as raw bytes. I am not using a framework or `net/http` here, because that would skip the part I am trying to learn.

## Purpose

I want to understand the full path from a socket to a parsed request, by writing the code myself.

I do not want another RFC summary or a blog post I forget a week later. I want something I can step through in a debugger when the buffer is only half full, when a header is missing a colon, or when the body is longer than it claimed to be. That is the kind of mess that actually teaches me how this works.

If I cannot explain why curl gets `200 OK` from a program I wrote, I do not understand the protocol yet.

## Intent

This is a learning project, so I am keeping it small.

I parse HTTP/1.1 myself, incrementally, because a real connection shows up in pieces. I keep tests next to the parser, because I do not trust a parser I have never tried to break. I am not hiding behind `net/http`. I am also not turning this into a product. If a feature does not teach me something about the wire, it does not belong here.

The UDP sender is here because I wanted to feel the difference between a stream and a datagram. They use the same port, but they are not the same contract, and I needed that contrast in my hands.

## What it does

- Parses `METHOD /path HTTP/1.1` from a byte stream
- Parses headers one line at a time (`Name: value`, CRLF)
- Reads the body from `Content-Length`
- Accepts TCP and prints the parsed request
- Writes a hardcoded `200 OK` back over the same connection

## Layout

```
.
├── cmd/
│   ├── httpserver/     HTTP server on :42069
│   ├── tcplistener/    Accept TCP, parse, print the request
│   └── udpsender/      Send stdin lines over UDP to :42069
├── internal/
│   ├── headers/        Header map + incremental parser
│   ├── request/        Request line + body state machine
│   └── server/         TCP accept loop and response writer
└── main.go             Empty entrypoint. Use the cmds.
```

## Requirements

Go 1.22+. Module path is `github.com/eclipsu/tcp2http`.

## Run

Parse incoming HTTP on TCP:

```bash
go run ./cmd/tcplistener
```

In another terminal:

```bash
printf 'GET / HTTP/1.1\r\nHost: localhost\r\n\r\n' | nc localhost 42069
```

Or:

```bash
curl -v http://localhost:42069/
```

Serve a response:

```bash
go run ./cmd/httpserver
```

Then:

```bash
curl http://localhost:42069/
```

You should see `Hello World!`.

UDP sender (stdin to `localhost:42069`):

```bash
go run ./cmd/udpsender
```

## Test

```bash
go test ./...
```

Parser tests live in `internal/headers` and `internal/request`. I care about those more than the binaries.

## How the parser works

`RequestFromReader` fills a sliding buffer and feeds it to a state machine.

1. **init**: wait for the first `\r\n`, split into method / target / version
2. **headers**: parse one `Key: value` line until a blank line
3. **body**: if `Content-Length` is set, read that many bytes. Otherwise done.

If there is not enough data yet, parse returns `n == 0` and we read again. A socket does not hand you a complete request, so you have to assemble one.

## Port

Default bind is `42069`.

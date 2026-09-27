package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/eclipsu/tcp2http/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateDone    parserState = "done"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	for r.state != StateDone {
		if r.state == StateInit {
			read_line, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break
			}

			r.RequestLine = *read_line
			read += n
			r.state = StateHeaders
		} else if r.state == StateHeaders {
			n, done, err := r.Headers.Parse(data[read:])

			if err != nil {
				return 0, err
			}

			if n == 0 && !done {
				break
			}

			read += n
			if done {
				r.state = StateBody
			}
		} else if r.state == StateBody {
			content_length := r.Headers.Get("content-length")
			if content_length == "" {
				r.state = StateDone
			} else {
				length, err := strconv.Atoi(content_length)
				if err != nil {
					return 0, fmt.Errorf("invalid content-length: %w", err)
				}

				remaining := data[read:]
				r.Body = append(r.Body, remaining...)
				read += len(remaining)

				if len(r.Body) > length {
					return 0, errors.New("body longer then content-length")
				}

				if len(r.Body) == length {
					r.state = StateDone
				} else {
					break
				}

			}
		}
	}
	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufIdx := 0

	for !request.done() {
		n, err := reader.Read(buf[bufIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		bufIdx += n

		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}

	if !request.done() {
		return nil, errors.New("unexpected EOF: incomplete request")
	}

	return request, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	lineEnd := bytes.Index(data, []byte("\r\n"))
	if lineEnd == -1 {
		return nil, 0, nil
	}

	line := data[:lineEnd]
	read := lineEnd + len("\r\n")

	parts := bytes.Split(line, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("request line has wrong number of parts: %d", len(parts))
	}

	method := parts[0]
	for _, char := range method {
		if char < 'A' || char > 'Z' {
			return nil, 0, fmt.Errorf("invalid method: %s", method)
		}
	}

	versionParts := bytes.Split(parts[2], []byte("/"))
	if len(versionParts) != 2 || string(versionParts[0]) != "HTTP" || string(versionParts[1]) != "1.1" {
		return nil, 0, fmt.Errorf("invalid HTTP version: %s", parts[2])
	}

	return &RequestLine{
		Method:        string(method),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(versionParts[1]),
	}, read, nil
}

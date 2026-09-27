package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(name string) string {
	return h[strings.ToLower(name)]
}

func (h Headers) Set(key string, value string) {

	if existing, ok := h[strings.ToLower(key)]; ok {
		h[strings.ToLower(key)] = existing + ", " + value
	} else {
		h[strings.ToLower(key)] = value
	}

}

var allowed = map[byte]bool{}

func init() {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!#$%&'*+-.^_`|~"
	for _, char := range []byte(chars) {
		allowed[char] = true
	}
}

func isValidKey(key []byte) bool {
	if len(key) == 0 {
		return false
	}

	for _, char := range key {
		if !allowed[char] {
			return false
		}
	}

	return true

}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	line := data[:idx]

	colon := bytes.IndexByte(line, ':')
	if colon == -1 {
		return 0, false, fmt.Errorf("malformed header: no colon")
	}

	key := line[:colon]
	if !isValidKey(key) {
		return 0, false, fmt.Errorf("invalid header key %q", key)
	}

	value := bytes.TrimSpace(line[colon+1:])

	h.Set(string(key), string(value))

	return idx + 2, false, nil
}

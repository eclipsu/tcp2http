package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"]) // lowercase now
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host:   localhost:42069   \r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 28, n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["host"] = "localhost:42069"
		data := []byte("User-Agent: curl/7.81.0\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, "curl/7.81.0", headers["user-agent"]) // lowercase now
		assert.Equal(t, 25, n)
		assert.False(t, done)
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		n, done, err := headers.Parse([]byte("\r\n"))
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.True(t, done)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		n, done, err := headers.Parse([]byte("       Host: localhost:42069\r\n\r\n"))
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid space before colon", func(t *testing.T) {
		headers := NewHeaders()
		n, done, err := headers.Parse([]byte("Host : localhost:42069\r\n\r\n"))
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Incomplete data needs more", func(t *testing.T) {
		headers := NewHeaders()
		n, done, err := headers.Parse([]byte("Host: localhost"))
		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Empty(t, headers)
	})

	t.Run("Only parses one header per call", func(t *testing.T) {
		headers := NewHeaders()
		n, done, err := headers.Parse([]byte("Host: a\r\nAccept: b\r\n\r\n"))
		require.NoError(t, err)
		assert.Equal(t, 9, n)
		assert.False(t, done)
		assert.Len(t, headers, 1)
	})

	// New for this lesson

	t.Run("Mixed case key is lowercased", func(t *testing.T) {
		headers := NewHeaders()
		_, _, err := headers.Parse([]byte("CoNtEnT-LeNgTh: 42\r\n\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "42", headers["content-length"])
		_, exists := headers["CoNtEnT-LeNgTh"]
		assert.False(t, exists) // original casing must not be stored
	})

	t.Run("Invalid character in key", func(t *testing.T) {
		headers := NewHeaders()
		n, done, err := headers.Parse([]byte("H©st: localhost:42069\r\n\r\n"))
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Empty(t, headers) // nothing stored on error
	})

	t.Run("Valid special characters in key", func(t *testing.T) {
		headers := NewHeaders()
		_, _, err := headers.Parse([]byte("X-Custom_Header.1~: ok\r\n\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "ok", headers["x-custom_header.1~"])
	})

	t.Run("Empty key", func(t *testing.T) {
		headers := NewHeaders()
		_, _, err := headers.Parse([]byte(": value\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("", func(t *testing.T) {
		headers := NewHeaders()
		_, _, err := headers.Parse([]byte(": value\r\n\r\n"))
		require.Error(t, err)
	})
	t.Run("Duplicate header with existing value", func(t *testing.T) {
		headers := NewHeaders()
		headers["set-person"] = "lane-loves-go" // lowercase key, like Parse stores it
		data := []byte("Set-Person: prime-loves-zig\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
		assert.Equal(t, 29, n)
		assert.False(t, done)
	})

	t.Run("Three duplicate headers in a row", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Set-Person: lane-loves-go\r\n" +
			"Set-Person: prime-loves-zig\r\n" +
			"Set-Person: tj-loves-ocaml\r\n" +
			"\r\n")

		for {
			n, done, err := headers.Parse(data)
			require.NoError(t, err)
			data = data[n:]
			if done {
				break
			}
			require.NotZero(t, n) // guards against an infinite loop
		}

		assert.Equal(t,
			"lane-loves-go, prime-loves-zig, tj-loves-ocaml",
			headers["set-person"])
		assert.Len(t, headers, 1)
	})

	t.Run("Duplicate keys with different casing combine", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Set-Person: a\r\nSET-PERSON: b\r\n\r\n")

		n, _, err := headers.Parse(data)
		require.NoError(t, err)
		_, _, err = headers.Parse(data[n:])
		require.NoError(t, err)

		assert.Equal(t, "a, b", headers["set-person"])
	})

	t.Run("Different keys are not combined", func(t *testing.T) {
		headers := NewHeaders()
		headers["host"] = "localhost:42069"
		_, _, err := headers.Parse([]byte("Accept: */*\r\n\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, "*/*", headers["accept"])
	})
}

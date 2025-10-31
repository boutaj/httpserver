package request

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestFromReader(t *testing.T) {
	t.Run("valid request line minimal path", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
			numBytesPerRead: 3,
		}

		req, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, "GET", req.RequestLine.Method)
		assert.Equal(t, "/", req.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", req.RequestLine.HttpVersion)
	})

	t.Run("valid request line with longer path", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
			numBytesPerRead: 1,
		}

		req, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, "GET", req.RequestLine.Method)
		assert.Equal(t, "/coffee", req.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", req.RequestLine.HttpVersion)
	})

	t.Run("invalid method character", func(t *testing.T) {
		reader := strings.NewReader("G@T / HTTP/1.1\r\n")

		_, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidMethodToken)
	})

	t.Run("invalid http version", func(t *testing.T) {
		reader := strings.NewReader("GET / HTTP/1\r\n")

		_, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidHTTPVersion)
	})

	t.Run("unexpected eof before request line complete", func(t *testing.T) {
		reader := strings.NewReader("GET / HTTP/1.1")

		_, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})

	t.Run("request line too large", func(t *testing.T) {
		overSizedPath := "/" + strings.Repeat("a", maxRequestLineBytes)
		line := "GET " + overSizedPath + " HTTP/1.1\r\n"
		reader := strings.NewReader(line)

		_, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.ErrorIs(t, err, errRequestLineTooLarge)
	})

	t.Run("propagates reader error", func(t *testing.T) {
		reader := errorReader{err: errors.New("read failure")}

		_, err := RequestFromReader(reader)
		require.Error(t, err)
		assert.EqualError(t, err, "read failure")
	})
}

func TestParseRequestLine(t *testing.T) {
	t.Run("valid tokens", func(t *testing.T) {
		line := []byte("GET /coffee HTTP/1.1")

		r, err := parseRequestLine(line)
		require.NoError(t, err)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/coffee", r.RequestTarget)
		assert.Equal(t, "1.1", r.HttpVersion)
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := parseRequestLine([]byte("GET /coffee"))
		require.Error(t, err)
		assert.ErrorIs(t, err, errRequestLineFormat)
	})

	t.Run("invalid method token", func(t *testing.T) {
		_, err := parseRequestLine([]byte("G@T /coffee HTTP/1.1"))
		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidMethodToken)
	})

	t.Run("invalid version numeric", func(t *testing.T) {
		_, err := parseRequestLine([]byte("GET /coffee HTTP/one.two"))
		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidHTTPVersion)
	})
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call.
// It is useful for simulating reading a variable number of bytes per chunk from a network connection.
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

type errorReader struct {
	err error
}

func (er errorReader) Read([]byte) (int, error) {
	return 0, er.err
}

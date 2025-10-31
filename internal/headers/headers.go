package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var (
	crlf                  = []byte("\r\n")
	errMalformedHeader    = errors.New("malformed header line")
	errInvalidHeaderToken = errors.New("invalid header name")
)

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	h.headers[strings.ToLower(name)] = value
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	index := bytes.Index(data, crlf)
	if index == -1 {
		return 0, false, nil
	}

	if index == 0 {
		return len(crlf), true, nil
	}

	line := data[:index]
	name, value, err := parseHeaderLine(line)
	if err != nil {
		return 0, false, err
	}

	h.Set(name, value)
	return index + len(crlf), false, nil
}

func parseHeaderLine(line []byte) (string, string, error) {
	colon := bytes.IndexByte(line, ':')
	if colon <= 0 {
		return "", "", errMalformedHeader
	}

	name := line[:colon]
	if !isValidHeaderName(name) {
		return "", "", fmt.Errorf("%w: %q", errInvalidHeaderToken, name)
	}

	value := bytes.TrimLeft(line[colon+1:], " \t")
	value = bytes.TrimRight(value, " \t")

	return string(name), string(value), nil
}

func isValidHeaderName(name []byte) bool {
	if len(name) == 0 {
		return false
	}

	for _, b := range name {
		if b > 127 {
			return false
		}
		if !isTChar(b) {
			return false
		}
	}

	return true
}

func isTChar(b byte) bool {
	switch {
	case b >= '0' && b <= '9':
		return true
	case b >= 'A' && b <= 'Z':
		return true
	case b >= 'a' && b <= 'z':
		return true
	}

	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}

package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type parserState int

const (
	parserInit parserState = iota
	parserDone
)

const (
	initialBufferSize   = 1024
	maxRequestLineBytes = 8192
	httpVersionPrefix   = "HTTP/"
)

var (
	crlf                   = []byte("\r\n")
	errRequestLineFormat   = errors.New("invalid request line format")
	errInvalidHTTPVersion  = errors.New("invalid http version")
	errInvalidMethodToken  = errors.New("invalid method token")
	errRequestLineTooLarge = errors.New("request line exceeds maximum length")
)

type Request struct {
	RequestLine RequestLine
	state       parserState
}

type RequestLine struct {
	HttpVersion   string // 1.1
	RequestTarget string // /coffee
	Method        string // GET
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == parserDone {
		return 0, nil
	}

	line, consumed := readLine(data)
	if consumed == 0 {
		if len(data) >= maxRequestLineBytes {
			return 0, errRequestLineTooLarge
		}
		return 0, nil
	}

	requestLine, err := parseRequestLine(line)
	if err != nil {
		return 0, err
	}

	r.RequestLine = requestLine
	r.state = parserDone

	return consumed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{state: parserInit}

	buf := make([]byte, initialBufferSize)
	bufLen := 0

	for request.state != parserDone {
		if bufLen == len(buf) {
			if len(buf) >= maxRequestLineBytes {
				return nil, errRequestLineTooLarge
			}
			nextSize := len(buf) * 2
			if nextSize > maxRequestLineBytes {
				nextSize = maxRequestLineBytes
			}
			grown := make([]byte, nextSize)
			copy(grown, buf[:bufLen])
			buf = grown
		}

		n, err := reader.Read(buf[bufLen:])
		if n > 0 {
			bufLen += n
			consumed, parseErr := request.parse(buf[:bufLen])
			if parseErr != nil {
				return nil, parseErr
			}
			if consumed > 0 {
				copy(buf, buf[consumed:bufLen])
				bufLen -= consumed
			}
		}

		if err != nil {
			if err == io.EOF {
				if request.state != parserDone {
					return nil, io.ErrUnexpectedEOF
				}
				break
			}
			return nil, err
		}
	}

	return request, nil
}

func parseRequestLine(line []byte) (RequestLine, error) {
	fields := bytes.Fields(line)
	if len(fields) != 3 {
		return RequestLine{}, errRequestLineFormat
	}

	method := fields[0]
	if !isToken(method) {
		return RequestLine{}, fmt.Errorf("%w: %q", errInvalidMethodToken, method)
	}

	version := fields[2]
	if !bytes.HasPrefix(version, []byte(httpVersionPrefix)) {
		return RequestLine{}, fmt.Errorf("%w: %q", errInvalidHTTPVersion, version)
	}

	versionNumber := version[len(httpVersionPrefix):]
	if !isValidHTTPVersion(versionNumber) {
		return RequestLine{}, fmt.Errorf("%w: %q", errInvalidHTTPVersion, version)
	}

	return RequestLine{
		Method:        string(method),
		RequestTarget: string(fields[1]),
		HttpVersion:   string(versionNumber),
	}, nil
}

func readLine(data []byte) ([]byte, int) {
	index := bytes.Index(data, crlf)
	if index == -1 {
		return nil, 0
	}
	return data[:index], index + len(crlf)
}

func isToken(token []byte) bool {
	if len(token) == 0 {
		return false
	}

	for _, b := range token {
		if b > 127 {
			return false
		}
		if !isTChar(b) {
			return false
		}
	}

	return true
}

func isValidHTTPVersion(version []byte) bool {
	if len(version) == 0 {
		return false
	}

	dotSeen := false
	for _, b := range version {
		if b == '.' {
			if dotSeen {
				return false
			}
			dotSeen = true
			continue
		}
		if b < '0' || b > '9' {
			return false
		}
	}

	return dotSeen
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

package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string // 1.1
	RequestTarget string // /coffee
	Method        string // GET
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	parsed, _, err := parseRequestLine(data)
	if err != nil {
		return nil, err
	}
	var rq Request
	rq.RequestLine = *parsed
	return &rq, nil
}

func parseRequestLine(line []byte) (*RequestLine, string, error) {
	sLine := string(line)
	index := strings.Index(sLine, "\r\n")

	if index == -1 {
		return nil, "", errors.New("failed to read request line")
	}

	rqline := sLine[:index]
	rest := sLine[index+2:]

	parts := strings.Split(string(rqline), " ")
	if len(parts) != 3 {
		return nil, rest, errors.New("invalid request line format")
	}
	method, target, version := parts[0], parts[1], parts[2]
	if !strings.HasPrefix(version, "HTTP/") {
		return nil, rest, errors.New("Invalid HTTP version : " + version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   strings.TrimPrefix(version, "HTTP/"),
	}, rest, nil
}

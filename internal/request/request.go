package request

import (
	"bytes"
	"errors"
	"io"
)

type parserState string
const (
	parserInit parserState = "initialized"
	parserDone parserState = "done"
)

type Request struct {
	RequestLine RequestLine
	state parserState
}

type RequestLine struct {
	HttpVersion   string // 1.1
	RequestTarget string // /coffee
	Method        string // GET
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
	outer:
		for {
			switch r.state {
				case parserInit:
					rl, n, err := parseRequestLine(data[read:])
					if err != nil {
						return 0, err
					}
					if n == 0 {
						break outer
					}
					r.RequestLine = *rl
					read += n

					r.state = parserDone
				case parserDone:
					break outer
			}
		}
	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{state: parserInit}

	buf := make([]byte, 1024)
	bufLen := 0
	for request.state != parserDone {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil	
}

func parseRequestLine(line []byte) (*RequestLine, int, error) {
	index := bytes.Index(line, []byte("\r\n"))

	if index == -1 {
		return nil, 0, nil
	}

	rqline := line[:index]
	rest := index + 2

	parts := bytes.Split(rqline, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, errors.New("invalid request line format")
	}
	method, target, version := parts[0], parts[1], parts[2]
	if !bytes.HasPrefix(version, []byte("HTTP/")) {
		return nil, 0, errors.New("Invalid HTTP version : " + string(version))
	}

	return &RequestLine{
		Method:        string(method),
		RequestTarget: string(target),
		HttpVersion:   string(version[len("HTTP/"):]),
	}, rest, nil
}

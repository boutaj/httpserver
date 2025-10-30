package headers

import (
	"bytes"
	"errors"
	"regexp"
)

type Headers map[string]string

func NewHeaders() (Headers) {
	header := Headers{}
	return header
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	CRLF  := []byte("\r\n")
	index := bytes.Index(data, CRLF)

	if index == -1 {
		return 0, false, nil
	}

	if index == 0 {
		return len(CRLF), true, nil
	}

	headerRe := regexp.MustCompile(`^[ \t]*([^\s:]+):[ \t]*(.*?)\s*$`)
	line     := bytes.Trim(data[:index], " ")
	parts    := headerRe.FindSubmatch(line)
	if parts == nil {
		return 0, false, errors.New("malformed header line")
	}

	h[string(parts[1])] = string(parts[2])

	return index + len(CRLF), false, nil
}
package parser

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type UmimplementedError struct {
	sigil byte
}

func (e *UmimplementedError) Error() string {
	return "Unimplemented parser for type: " + string(e.sigil)
}

type InvalidInputError struct {
	msg string
}

func (e *InvalidInputError) Error() string {
	return e.msg
}

type Parser interface {
	// Parse takes a byte slice, and an index to start parsing from.
	// It returns (nil, i, err), where `e` is an error, if the input is invalid. In this case, `i` is the index where the error was detected, and it may be possible to resume parsing from that index.
	// It returns (nil, i, nil) if the input is incomplete, and it is possible to resume parsing from that index.
	// It returns (resp.RESP, i, nil) if the input is complete, and the parsed RESP object is returned.
	Parse(bs []byte, start int) (resp.RESP, int, error)
}

type ToplevelParser struct {
	subParser Parser
}

func (p *ToplevelParser) Parse(bs []byte, start int) (resp.RESP, int, error) {
	fmt.Println("ToplevelParser.Parse")

	if p.subParser != nil {
		r, i, err := p.subParser.Parse(bs, start)
		// handle error
		if err != nil {
			return r, i, err
		}
		// handle complete parsing
		if r != nil {
			p.subParser = nil
		}
		return r, i, nil
	}
	if start >= len(bs) {
		return nil, start, nil
	}
	switch bs[start] {
	case ':':
		p.subParser = newIntegerParser()
		return p.subParser.Parse(bs, start+1)
	case '$':
		p.subParser = newBulkStringParser()
		return p.subParser.Parse(bs, start+1)
	case '*':
		p.subParser = newArrayParser()
		return p.subParser.Parse(bs, start+1)
	case '_':
		p.subParser = newNullParser()
		return p.subParser.Parse(bs, start+1)
	case '#':
		p.subParser = newBooleanParser()
		return p.subParser.Parse(bs, start+1)
	case '!':
		p.subParser = newBulkErrorParser()
		return p.subParser.Parse(bs, start+1)
	default:
		return nil, start, &UmimplementedError{bs[start]}
	}
}

func NewParser() Parser {
	return &ToplevelParser{}
}

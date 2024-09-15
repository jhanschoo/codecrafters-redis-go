package parser

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type internalIntParser struct {
	buf        []byte
	terminated bool
}

func newInternalIntParser() *internalIntParser {
	return &internalIntParser{make([]byte, 0, 10), false}
}

func (p *internalIntParser) Parse(bs []byte, start int) (resp.RESP, int, error) {
	fmt.Println("internalIntParser.Parse")
	i := start
	for ; i < len(bs); i++ {
		if p.terminated {
			if bs[i] == '\n' {
				value, err := strconv.ParseInt(string(p.buf), 10, 64)
				if err != nil {
					p.buf = p.buf[:0]
					return nil, i + 1, &InvalidInputError{fmt.Sprintf("Invalid input while parsing integer: at index %v, %v", i, err)}
				}
				fmt.Println("internalIntParser.Parse: value", value)
				return &resp.RESPInteger{Value: value}, i + 1, nil
			}
			return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing integer: at index %v, expected byte('\\n'), got %v", i, bs[i])}
		}
		if bs[i] == '\r' {
			p.terminated = true
			continue
		}

		if bs[i] >= '0' && bs[i] <= '9' {
			p.buf = append(p.buf, bs[i])
			continue
		}

		return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing integer: at index %v, expected byte('0'-'9') or byte('\\r'), got %v", i, bs[i])}
	}
	return nil, i, nil
}

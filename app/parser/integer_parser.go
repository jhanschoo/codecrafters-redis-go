package parser

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type integerParser struct {
	buf        []byte
	allowSign  bool
	terminated bool
}

func newIntegerParser() *integerParser {
	return &integerParser{make([]byte, 0, 20), true, false}
}

func (p *integerParser) Parse(bs []byte, start int) (resp.RESP, int, error) {
	fmt.Println("integerParser.Parse")
	i := start
	for ; i < len(bs); i++ {
		if p.terminated {
			if bs[i] == '\n' {
				value, err := strconv.ParseInt(string(p.buf), 10, 64)
				if err != nil {
					p.buf = p.buf[:0]
					return nil, i + 1, &InvalidInputError{fmt.Sprintf("Invalid input while parsing integer: at index %v, %v", i, err)}
				}
				return &resp.RESPInteger{Value: value}, i + 1, nil
			}
			return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing integer: at index %v, expected byte('\\n'), got %v", i, bs[i])}
		}
		if bs[i] == '\r' {
			p.terminated = true
			continue
		}
		if bs[i] == '+' || bs[i] == '-' {
			if !p.allowSign {
				return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing integer: at index %v, expected byte('0'-'9') or byte('\\r'), got %v", i, bs[i])}
			}
			p.buf = append(p.buf, bs[i])
			p.allowSign = false
			continue
		}

		if bs[i] >= '0' && bs[i] <= '9' {
			p.buf = append(p.buf, bs[i])
			p.allowSign = false
			continue
		}

		return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing integer: at index %v, expected byte('0'-'9') or byte('\\r') (or perhaps byte('+' or '-')), got %v", i, bs[i])}
	}
	return nil, i, nil
}

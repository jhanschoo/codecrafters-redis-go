package parser

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type booleanParser struct {
	B bool
	// `N` is the number of the serialized null bytes we have seen so far.
	N int
}

func newBooleanParser() *booleanParser {
	return &booleanParser{false, 1}
}

func (p *booleanParser) Parse(bs []byte, start int) (resp.RESP, int, error) {
	fmt.Println("booleanParser.Parse")
	for i := start; i < len(bs); i++ {
		switch p.N {
		case 1:
			switch bs[i] {
			case 't':
				p.B = true
			case 'f':
				p.B = false
			default:
				return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing boolean: at index %v, expected byte('t') or byte('f'), got %v", i, bs[i])}
			}
		case 2:
			if bs[i] != '\r' {
				return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing boolean: at index %v, expected byte('\\r'), got %v", i, bs[i])}
			}
		case 3:
			if bs[i] != '\n' {
				return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing boolean: at index %v, expected byte('\\n'), got %v", i, bs[i])}
			}
			fmt.Printf("booleanParser.Parse: %v\n", p.B)
			return &resp.RESPBoolean{Value: p.B}, i + 1, nil
		}
		p.N++
	}
	return nil, len(bs), nil
}

package parser

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type nullParser struct {
	crParsed bool
}

func newNullParser() *nullParser {
	return &nullParser{false}
}

func (p *nullParser) Parse(bs []byte, start int) (resp.RESP, int, error) {
	fmt.Println("nullParser.Parse")
	i := start
	for ; i < len(bs); i++ {
		if !p.crParsed {
			if bs[i] != '\r' {
				return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing null: at index %v, expected byte('\\r'), got %v", i, bs[i])}
			}
			p.crParsed = true
			continue
		}
		if bs[i] != '\n' {
			return nil, i, &InvalidInputError{fmt.Sprintf("Invalid input while parsing null: at index %v, expected byte('\\n'), got %v", i, bs[i])}
		}
		return &resp.RESPNull{}, i + 1, nil
	}
	return nil, i, nil
}

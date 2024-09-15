package parser

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type arrayParser struct {
	buf         []resp.RESP
	subParser   Parser
	arrayLength int64
	// `N` is the number of elements of the array proper we have parsed so far.
	N int64
}

func newArrayParser() *arrayParser {
	return &arrayParser{nil, newInternalIntParser(), -1, 0}
}

func (p *arrayParser) Parse(bs []byte, start int) (resp.RESP, int, error) {
	fmt.Println("arrayParser.Parse")
	// handle parsing of the length of the string
	if p.arrayLength == -1 {
		r, i, err := p.subParser.Parse(bs, start)
		// handle error
		if err != nil {
			return r, i, err
		}
		// handle incomplete parsing
		if r == nil {
			return nil, i, nil
		}
		// handle complete parsing
		p.subParser = NewParser()
		p.arrayLength = r.(*resp.RESPInteger).Value
		p.buf = make([]resp.RESP, 0, p.arrayLength)
		p.N = 0
		start = i
		fmt.Println("arrayParser.Parse: arrayLength", p.arrayLength)
	}
	// handle parsing of the string proper
	i := start
	for ; i < len(bs) && p.N < p.arrayLength; p.N++ {
		r, j, err := p.subParser.Parse(bs, i)
		// handle error
		if err != nil {
			return r, j, err
		}
		// handle incomplete parsing
		if r == nil {
			return nil, j, nil
		}
		// handle complete parsing
		p.buf = append(p.buf, r)
		fmt.Printf("arrayParser.Parse: index %v parsed, at buffer index %v\n", p.N, j)
		i = j
		p.subParser = NewParser()
	}
	// handle incomplete parsing
	if p.N < p.arrayLength {
		return nil, i, nil
	}
	// handle complete parsing
	return &resp.RESPArray{Value: p.buf}, i, nil
}

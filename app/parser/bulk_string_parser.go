package parser

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bulkStringParser struct {
	buf          []byte
	subParser    Parser
	stringLength int64
	// `N` is the number of bytes of the string proper we have seen so far.
	N int64
}

func newBulkStringParser() *bulkStringParser {
	return &bulkStringParser{nil, newInternalIntParser(), -1, 0}
}

func (p *bulkStringParser) Parse(bs []byte, start int) (resp.RESP, int, error) {
	fmt.Println("bulkStringParser.Parse")
	// handle parsing of the length of the string
	if p.stringLength == -1 {
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
		p.subParser = newNullParser()
		p.stringLength = r.(*resp.RESPInteger).Value
		p.buf = make([]byte, 0, p.stringLength)
		p.N = 0
		start = i
		fmt.Println("bulkStringParser.Parse: stringLength", p.stringLength)
	}
	// handle parsing of the string proper
	i := start
	for ; i < len(bs) && p.N < p.stringLength; i, p.N = i+1, p.N+1 {
		p.buf = append(p.buf, bs[i])
	}
	// handle incomplete parsing
	if p.N < p.stringLength {
		return nil, i, nil
	}
	// handle parsing of the CRLF
	r, i, err := p.subParser.Parse(bs, i)
	// handle error
	if err != nil {
		return r, i, err
	}
	// handle incomplete parsing
	if r == nil {
		return nil, i, nil
	}
	// handle complete parsing
	fmt.Printf("bulkStringParser.Parse: bulk string parsed %d bytes\n", p.stringLength)
	return &resp.RESPBulkString{Value: p.buf}, i, nil
}

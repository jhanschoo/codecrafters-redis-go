package parser

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bulkErrorParser struct {
	subParser Parser
}

func newBulkErrorParser() *bulkErrorParser {
	return &bulkErrorParser{newBulkStringParser()}
}

func (p *bulkErrorParser) Parse(bs []byte, start int) (resp.RESP, int, error) {
	fmt.Println("bulkErrorParser.Parse")
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
	fmt.Println("bulkErrorParser.Parse: bulk error parsed")
	return &resp.RESPBulkError{Value: r.(*resp.RESPBulkString).Value}, i, nil
}

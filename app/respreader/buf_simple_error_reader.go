package respreader

import (
	"bufio"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufSimpleErrorReader struct {
	br *bufSimpleStringReader
}

func newBufSimpleErrorReader(br *bufio.Reader) Reader {
	return &bufSimpleErrorReader{newInternalBufSimpleStringReader(br)}
}

func (rr *bufSimpleErrorReader) ReadRESP() (resp.RESP, error) {
	fmt.Println("bufSimpleErrorReader.ReadRESP, delegating to bufSimpleStringReader")
	rstr, err := rr.br.readRESPSimpleString()
	if rstr == nil {
		return nil, err
	}
	return &resp.RESPSimpleError{Value: rstr.Value}, err
}

package respreader

import (
	"bufio"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufSimpleErrorReader struct {
	br *bufSimpleStringReader
}

var _ Reader = (*bufSimpleErrorReader)(nil)

func newBufSimpleErrorReader(br *bufio.Reader) Reader {
	return &bufSimpleErrorReader{newInternalBufSimpleStringReader(br)}
}

func (rr *bufSimpleErrorReader) ReadRESP() (resp.RESP, error) {
	rstr, err := rr.br.readRESPSimpleString()
	if rstr == nil {
		return nil, err
	}
	return &resp.RESPSimpleError{Value: rstr.Value}, err
}

package respreader

import (
	"bufio"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufSimpleStringReader struct {
	br  *bufio.Reader
	buf []byte
}

var _ Reader = (*bufSimpleStringReader)(nil)

func newBufSimpleStringReader(br *bufio.Reader) Reader {
	return newInternalBufSimpleStringReader(br)
}

func newInternalBufSimpleStringReader(br *bufio.Reader) *bufSimpleStringReader {
	return &bufSimpleStringReader{br, make([]byte, 0, 4096)}
}

func (rr *bufSimpleStringReader) ReadRESP() (resp.RESP, error) {
	return rr.readRESPSimpleString()
}

func (rr *bufSimpleStringReader) readRESPSimpleString() (*resp.RESPSimpleString, error) {
	bs, err := rr.br.ReadBytes('\n')
	rr.buf = append(rr.buf, bs...)
	if err != nil {
		return nil, err
	}
	v, err := stripTerminator(rr.buf)
	if err != nil {
		return nil, err
	}
	return &resp.RESPSimpleString{Value: string(v)}, nil
}

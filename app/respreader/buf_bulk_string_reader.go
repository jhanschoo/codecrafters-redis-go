package respreader

import (
	"bufio"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

// expose the struct for nonstandard uses
type BufBulkStringReader struct {
	br           *bufio.Reader
	lengthReader *bufIntegerReader
	length       int64
	buf          []byte
	n            int64
}

func NewBufBulkStringReader(br *bufio.Reader) *BufBulkStringReader {
	return &BufBulkStringReader{br, newInternalBufIntegerReader(br), -2, nil, 0}
}

var _ Reader = (*BufBulkStringReader)(nil)

func newBufBulkStringReader(br *bufio.Reader) Reader {
	return NewBufBulkStringReader(br)
}

func (rr *BufBulkStringReader) ReadRESPUnterminated() (resp.RESP, error) {
	if rr.buf == nil {
		rint, err := rr.lengthReader.readRESPInteger()
		if err != nil {
			return nil, err
		}
		if rint.Value < 0 {
			return nil, ErrorNegativeLength
		}
		rr.length = rint.Value
		rr.lengthReader = nil
		rr.buf = make([]byte, rr.length)
		rr.n = 0
	}
	for rr.n < rr.length {
		n, err := rr.br.Read(rr.buf[rr.n:])
		rr.n += int64(n)
		if err != nil {
			return nil, err
		}
	}
	return &resp.RESPBulkString{Value: string(rr.buf)}, nil
}

func (rr *BufBulkStringReader) ReadRESP() (resp.RESP, error) {
	if rr.buf == nil {
		rint, err := rr.lengthReader.readRESPInteger()
		if err != nil {
			return nil, err
		}
		if rint.Value < 0 {
			return nil, ErrorNegativeLength
		}
		rr.length = rint.Value
		rr.lengthReader = nil
		rr.buf = make([]byte, rr.length+2)
		rr.n = 0
	}
	for rr.n < rr.length+2 {
		n, err := rr.br.Read(rr.buf[rr.n:])
		rr.n += int64(n)
		if err != nil {
			return nil, err
		}
	}

	buf, err := stripTerminator(rr.buf)
	if err != nil {
		return nil, err
	}
	return &resp.RESPBulkString{Value: string(buf)}, nil
}

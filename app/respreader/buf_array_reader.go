package respreader

import (
	"bufio"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufArrayReader struct {
	br            *bufio.Reader
	lengthReader  *bufIntegerReader
	length        int64
	elementReader *bufPayloadReader
	buf           []resp.RESP
}

var _ Reader = (*bufArrayReader)(nil)

func newBufArrayReader(br *bufio.Reader) Reader {
	return &bufArrayReader{br, newInternalBufIntegerReader(br), -2, nil, nil}
}

func (rr *bufArrayReader) ReadRESP() (resp.RESP, error) {
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
		rr.buf = make([]resp.RESP, 0, rr.length)
	}
	for len(rr.buf) < int(rr.length) {
		if rr.elementReader == nil {
			rr.elementReader = newInternalBufPayloadReader(rr.br)
		}
		robj, err := rr.elementReader.ReadRESP()
		if err != nil {
			return nil, err
		}
		rr.buf = append(rr.buf, robj)
	}
	return &resp.RESPArray{Value: rr.buf}, nil
}

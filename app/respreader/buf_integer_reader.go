package respreader

import (
	"bufio"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufIntegerReader struct {
	br       *bufio.Reader
	ssReader *bufSimpleStringReader
}

var _ Reader = (*bufIntegerReader)(nil)

func newBufIntegerReader(br *bufio.Reader) Reader {
	return newInternalBufSimpleStringReader(br)
}

func newInternalBufIntegerReader(br *bufio.Reader) *bufIntegerReader {
	return &bufIntegerReader{br, newInternalBufSimpleStringReader(br)}
}

func (rr *bufIntegerReader) ReadRESP() (resp.RESP, error) {
	return rr.readRESPInteger()
}

func (rr *bufIntegerReader) readRESPInteger() (*resp.RESPInteger, error) {
	rstr, err := rr.ssReader.readRESPSimpleString()
	if err != nil {
		return nil, err
	}
	i, err := strconv.ParseInt(string(rstr.Value), 10, 64)
	if err != nil {
		return nil, err
	}
	return &resp.RESPInteger{Value: i}, nil
}

package respreader

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufBulkStringReader struct {
	br           *bufio.Reader
	lengthReader *bufIntegerReader
	length       int64
	buf          []byte
	n            int64
}

func newBufBulkStringReader(br *bufio.Reader) Reader {
	return &bufBulkStringReader{br, newInternalBufIntegerReader(br), -2, nil, 0}
}

func (rr *bufBulkStringReader) ReadRESP() (resp.RESP, error) {
	if rr.buf == nil {
		fmt.Println("bufBulkStringReader.ReadRESP: parsing length")
		fmt.Println("bufBulkStringReader.ReadRESP: delegating to bufIntegerReader")
		rint, err := rr.lengthReader.readRESPInteger()
		if err != nil {
			return nil, err
		}
		if rint.Value < 0 {
			fmt.Println("bufBulkStringReader.ReadRESP: error negative length")
			return nil, errors.New("negative length")
		}
		rr.length = rint.Value
		rr.lengthReader = nil
		rr.buf = make([]byte, rr.length+2)
		rr.n = 0
	}
	for rr.n < rr.length+2 {
		fmt.Println("bufBulkStringReader.ReadRESP: reading bulk string")
		n, err := rr.br.Read(rr.buf[rr.n:])
		rr.n += int64(n)
		if err != nil {
			fmt.Println("bufBulkStringReader.ReadRESP: error", err)
			return nil, err
		}
	}

	buf, err := stripTerminator(rr.buf)
	if err != nil {
		return nil, err
	}
	return &resp.RESPBulkString{Value: string(buf)}, nil
}

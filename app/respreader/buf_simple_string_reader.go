package respreader

import (
	"bufio"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufSimpleStringReader struct {
	br  *bufio.Reader
	buf []byte
}

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
	fmt.Println("bufSimpleStringReader.readRESPSimpleString")
	bs, err := rr.br.ReadBytes('\n')
	rr.buf = append(rr.buf, bs...)
	if err != nil {
		fmt.Println("bufSimpleStringReader.readRESPSimpleString: error", err)
		return nil, err
	}
	v, err := stripTerminator(rr.buf)
	if err != nil {
		fmt.Println("simpleStringParser.Parse: error", err)
		return nil, err
	}
	return &resp.RESPSimpleString{Value: string(v)}, nil
}

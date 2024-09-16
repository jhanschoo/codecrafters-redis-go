package respreader

import (
	"bufio"
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufIntegerReader struct {
	br       *bufio.Reader
	ssReader *bufSimpleStringReader
}

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
	fmt.Println("bufIntegerReader.ReadRESPInteger: delegating to bufSimpleStringReader")
	rstr, err := rr.ssReader.readRESPSimpleString()
	if err != nil {
		return nil, err
	}
	fmt.Println("bufIntegerReader.readRESPInteger: parsing integer")
	i, err := strconv.ParseInt(string(rstr.Value), 10, 64)
	if err != nil {
		fmt.Println("bufIntegerReader.readRESPInteger: error", err)
		return nil, err
	}
	return &resp.RESPInteger{Value: i}, nil
}

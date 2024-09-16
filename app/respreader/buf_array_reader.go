package respreader

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufArrayReader struct {
	br            *bufio.Reader
	lengthReader  *bufIntegerReader
	length        int64
	elementReader *bufPayloadReader
	buf           []resp.RESP
}

func newBufArrayReader(br *bufio.Reader) Reader {
	return &bufArrayReader{br, newInternalBufIntegerReader(br), -2, nil, nil}
}

func (rr *bufArrayReader) ReadRESP() (resp.RESP, error) {
	if rr.buf == nil {
		fmt.Println("bufArrayReader.ReadRESP: parsing length")
		fmt.Println("bufArrayReader.ReadRESP: delegating to bufIntegerReader")
		rint, err := rr.lengthReader.readRESPInteger()
		if err != nil {
			return nil, err
		}
		if rint.Value < 0 {
			fmt.Println("bufArrayReader.ReadRESP: error negative length")
			return nil, errors.New("negative length")
		}
		rr.length = rint.Value
		rr.lengthReader = nil
		rr.buf = make([]resp.RESP, 0, rr.length)
	}
	for len(rr.buf) < int(rr.length) {
		if rr.elementReader == nil {
			rr.elementReader = newInternalBufPayloadReader(rr.br, true)
		}
		robj, err := rr.elementReader.ReadRESP()
		if err != nil {
			return nil, err
		}
		rr.buf = append(rr.buf, robj)
	}
	fmt.Println("bufArrayReader.ReadRESP: returning array of", len(rr.buf), "elements")
	return &resp.RESPArray{Value: rr.buf}, nil
}

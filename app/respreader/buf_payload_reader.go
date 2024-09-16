package respreader

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufPayloadReader struct {
	br         *bufio.Reader
	subReader  Reader
	isInternal bool
}

func newInternalBufPayloadReader(br *bufio.Reader, isInternal bool) *bufPayloadReader {
	return &bufPayloadReader{br, nil, isInternal}
}

var subreaderMap = map[byte]func(br *bufio.Reader) Reader{
	'+': newBufSimpleStringReader,
	'-': newBufSimpleErrorReader,
	':': newBufIntegerReader,
	'$': newBufBulkStringReader,
	'*': newBufArrayReader,
}

func (rr *bufPayloadReader) ReadRESP() (resp.RESP, error) {
	fmt.Println("bufPayloadReader.ReadRESP")
	if rr.subReader == nil {
		fmt.Println("bufPayloadReader.ReadRESP: not in intermediate state, reading type byte")
		b, err := rr.br.ReadByte()
		if err != nil {
			fmt.Println("bufPayloadReader.ReadRESP: error", err, "(parsing type byte restarted)")
			if rr.isInternal {
				err = &readerError{err, false}
			}
			return nil, err
		}
		fmt.Println("bufPayloadReader.ReadRESP: type byte", b)
		subReaderCreator, ok := subreaderMap[b]
		if !ok {
			fmt.Println("bufPayloadReader.ReadRESP: error unimplemented type", b)
			err = errors.New("bufPayloadReader.ReadRESP: unimplemented type")
			if rr.isInternal {
				err = &readerError{err, false}
			}
			return nil, err
		}
		rr.subReader = subReaderCreator(rr.br)
	}
	ret, err := rr.subReader.ReadRESP()
	if err != nil {
		v, ok := err.(*readerError)
		shouldRestart := true
		if ok {
			shouldRestart = v.shouldRestart
			err = v.Err
		}
		if shouldRestart && err != io.ErrNoProgress {
			fmt.Println("bufPayloadReader.ReadRESP: error", err, "(parsing payload restarted to type byte)")
			rr.subReader = nil
		}
		if rr.isInternal {
			err = &readerError{err, false}
		}
		return nil, err
	}
	rr.subReader = nil
	return ret, nil
}

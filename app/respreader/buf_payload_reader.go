package respreader

import (
	"bufio"
	"errors"
	"io"
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type bufPayloadReader struct {
	br         *bufio.Reader
	subReader  Reader
	isInternal bool
}

var _ Reader = (*bufPayloadReader)(nil)

func newInternalBufPayloadReader(br *bufio.Reader) *bufPayloadReader {
	return &bufPayloadReader{br, nil, true}
}

var subreaderMap = map[byte]func(br *bufio.Reader) Reader{
	'+': newBufSimpleStringReader,
	'-': newBufSimpleErrorReader,
	':': newBufIntegerReader,
	'$': newBufBulkStringReader,
	'*': newBufArrayReader,
}

func (rr *bufPayloadReader) ReadRESP() (resp.RESP, error) {
	if rr.subReader == nil {
		b, err := rr.br.ReadByte()
		if err != nil {
			if rr.isInternal {
				err = &shouldNotRestartError{err}
			}
			return nil, err
		}
		subReaderCreator, ok := subreaderMap[b]
		if !ok {
			err = ErrorUnimplementedDataType
			if rr.isInternal {
				err = &shouldNotRestartError{err}
			}
			return nil, err
		}
		rr.subReader = subReaderCreator(rr.br)
	}
	ret, err := rr.subReader.ReadRESP()
	if err != nil {
		shouldRestart := true
		if errors.As(err, (*shouldNotRestartError)(nil)) {
			err = errors.Unwrap(err)
			shouldRestart = false
		}
		if shouldRestart && err != io.ErrNoProgress {
			log.Printf("bufPayloadReader.ReadRESP: error %v (parsing payload restarted to just before type byte)\n", err)
			rr.subReader = nil
		}
		if rr.isInternal {
			err = &shouldNotRestartError{err}
		}
		return nil, err
	}
	rr.subReader = nil
	return ret, nil
}

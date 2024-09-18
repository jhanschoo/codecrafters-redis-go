// Package respreader provides a Parser for the RESP protocol.
package respreader

import (
	"bufio"
	"errors"
	"io"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type Reader interface {
	// ReadRESP attempts to read a RESP object from an input source it was initialized with.
	// It returns (nil, err), where `e` is an error, if an encounter was encountered while reading. For example, some Reader implementations may return (nil, io.EOF) if the input is empty.
	// Except for when io.ErrNoProgress is returned, when an error is returned, where meaningful, Reader implementations should recover to before when it started parsing the current RESP object, which may be an element of an aggregate (note that the redis documentation may contain an error where bulk data types are categorized as aggregates in at least one instance). That is, the next byte is expected to be a type byte of another RESP object.
	// It returns (resp.RESP, nil) if the input is complete, and the parsed RESP object is returned. the bufio.Reader is advanced to the next byte after the parsed RESP object.
	ReadRESP() (resp.RESP, error)
}

// shouldNotRestartError is an error that wraps around underlying errors to inform parent readers their parsing process should NOT be restarted. This internal signalling enables the parser to recover from errors at well-defined points in the parsing process, without discarding the parse altogether. At the time of writing, only the payload reader uses this.
// shouldRestart is true if the parsing process should be restarted from the beginning of the current RESP object, except when the underlying error is io.ErrNoProgress.
type shouldNotRestartError struct {
	error
}

var _ error = (*shouldNotRestartError)(nil)

func (e *shouldNotRestartError) Unwrap() error {
	return e.error
}

func NewBufReader(is io.Reader) Reader {
	r := newInternalBufPayloadReader(bufio.NewReader(is))
	r.isInternal = false
	return r
}

var (
	ErrorUnimplementedDataType = errors.New("RESPReader: unimplemented type")
	ErrorNegativeLength        = errors.New("RESPReader: negative length")
	ErrorInvalidTerminator     = errors.New("RESPReader: invalid terminator")
)

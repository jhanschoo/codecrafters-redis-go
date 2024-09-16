package respreader

import (
	"errors"
)

func stripTerminator(bs []byte) ([]byte, error) {
	if len(bs) < 2 || bs[len(bs)-2] != '\r' || bs[len(bs)-1] != '\n' {
		err := errors.New("invalid input: expected terminator []byte(\"\\r\\n\")")
		return nil, err
	}
	return bs[0 : len(bs)-2], nil
}

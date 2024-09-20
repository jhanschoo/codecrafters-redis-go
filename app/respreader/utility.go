package respreader

import (
	"bufio"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func stripTerminator(bs []byte) ([]byte, error) {
	if len(bs) < 2 || bs[len(bs)-2] != '\r' || bs[len(bs)-1] != '\n' {
		return nil, ErrorInvalidTerminator
	}
	return bs[0 : len(bs)-2], nil
}

type BufferedRESPConnReader struct {
	net.Conn
	*bufio.Reader
	RESPReader Reader
}

func (r *BufferedRESPConnReader) ReadRESP() (resp.RESP, error) {
	return r.RESPReader.ReadRESP()
}

func NewBufferedRESPConnReader(conn net.Conn) *BufferedRESPConnReader {
	br := bufio.NewReader(conn)
	rr := NewBufReader(br)
	return &BufferedRESPConnReader{
		Conn:       conn,
		Reader:     br,
		RESPReader: rr,
	}
}

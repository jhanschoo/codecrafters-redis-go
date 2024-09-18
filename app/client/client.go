package client

import (
	"errors"
	"io"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

type Client struct {
	io.Writer
	respreader.Reader
}

func NewReplicaClient(replicaof string) (*Client, error) {
	masterAddrSlice := strings.Split(replicaof, " ")
	if len(masterAddrSlice) != 2 {
		return nil, errors.New("invalid input: expected replicaof to be in the format 'host port'")
	}
	return NewClient(strings.Join(masterAddrSlice, ":"))
}

func NewClient(serverAddr string) (*Client, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	rr := respreader.NewBufReader(conn)
	return &Client{
		Writer: conn,
		Reader: rr,
	}, nil
}

func (c *Client) Close() error {
	if closer, ok := c.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (c *Client) Do(req []string) (resp.RESP, error) {
	if _, err := c.Write([]byte(resp.ParseStringSlice(req).SerializeRESP())); err != nil {
		return nil, err
	}
	return c.ReadRESP()
}

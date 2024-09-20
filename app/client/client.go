package client

import (
	"errors"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

type Client struct {
	*respreader.BufferedRESPConnReader
}

func (c *Client) ReadRESP() (resp.RESP, error) {
	return c.RESPReader.ReadRESP()
}

func NewReplicaToMasterClient() (*Client, error) {
	replicaof := config.Get("replicaof")
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
	return FromConn(conn), nil
}

func FromConn(conn net.Conn) *Client {
	return &Client{
		respreader.NewBufferedRESPConnReader(conn),
	}
}

func (c *Client) Do(req []string) (resp.RESP, error) {
	s := resp.EncodeStringSlice(req).SerializeRESP()
	log.Println("client.Do: sending", strconv.Quote(s))
	if _, err := c.Write([]byte(s)); err != nil {
		return nil, err
	}
	return c.ReadRESP()
}

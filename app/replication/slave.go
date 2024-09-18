package replication

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var (
	respPing = resp.RESPArray{Value: []resp.RESP{&resp.RESPSimpleString{Value: "PING"}}}.SerializeRESP()
	respPong = resp.RESPSimpleString{Value: "PONG"}.SerializeRESP()
)

func InitializeSlave(replInfo *ReplicationInfo) error {
	replInfo.Role = "slave"

	masterAddrSlice := strings.Split(replInfo.Replicaof, " ")
	if len(masterAddrSlice) != 2 {
		return errors.New("invalid input: expected replicaof to be in the format 'host port'")
	}
	masterAddr := strings.Join(masterAddrSlice, ":")
	tcpAddr, err := net.ResolveTCPAddr("tcp", masterAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}

	err = performHandshakeAsSlave(conn, replInfo)
	return err
}

func performHandshakeAsSlave(conn net.Conn, replInfo *ReplicationInfo) error {
	brw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	if _, err := brw.Write([]byte(respPing)); err != nil {
		return err
	}
	if err := brw.Flush(); err != nil {
		return err
	}
	// Receive PONG
	res, err := brw.ReadSlice('\n')
	if err != nil {
		return err
	}
	if string(res) != respPong {
		return fmt.Errorf("expected %s, received %s", respPong, string(res))
	}
	// Send REPLCONF listening-port <port>
	// Receive OK
	// Send SYNC
	// Receive +FULLRESYNC <replid> <offset>
	// Receive DB data
	// Receive +CONTINUE
	return nil
}

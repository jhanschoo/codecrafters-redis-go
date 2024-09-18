package replication

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func InitializeSlave(replInfo *ReplicationInfo) error {
	log.Println("initializeSlave: started")
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
	if err := reqExpectingSimpleStringRes(brw, []string{"PING"}, "PONG"); err != nil {
		return err
	}
	port, _ := config.Get("port")
	if err := reqExpectingSimpleStringRes(brw, []string{"REPLCONF", "listening-port", port}, "OK"); err != nil {
		return err
	}
	if err := reqExpectingSimpleStringRes(brw, []string{"REPLCONF", "capa", "psync2"}, "OK"); err != nil {
		return err
	}
	return nil
}

func reqExpectingSimpleStringRes(brw *bufio.ReadWriter, req []string, expectedRes string) error {
	reqArrayValue := make([]resp.RESP, 0, len(req))
	for _, r := range req {
		reqArrayValue = append(reqArrayValue, &resp.RESPBulkString{Value: r})
	}
	reqBytes := []byte(resp.RESPArray{Value: reqArrayValue}.SerializeRESP())
	if _, err := brw.Write(reqBytes); err != nil {
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
	expectedResResp := resp.RESPSimpleString{Value: expectedRes}.SerializeRESP()
	if string(res) != expectedResResp {
		return fmt.Errorf("expected %s, received %s", expectedResResp, string(res))
	}
	return nil
}

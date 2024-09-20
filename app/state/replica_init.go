package state

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/rdbreader"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

func initializeReplica() {
	log.Println("initializeReplica: started")
	var err error
	state.Role = "slave"
	state.MasterReplid = "?"
	state.MasterReplOffset.Store(-1)
	state.MasterClient, err = client.NewReplicaToMasterClient()
	if err != nil {
		log.Fatalln("initializeReplica: failed to create master client", err)
	}

	err = initiateHandshakeAsReplica()
	if err != nil {
		log.Fatalln("initializeReplica: failed to perform handshake", err)
	}
	log.Println("initializeReplica: handshake complete")

	err = performInitialSync()
	if err != nil {
		log.Fatalln("initializeReplica: failed to perform initial sync", err)
	}
	log.Println("initializeReplica: initial sync complete")
}

func initiateHandshakeAsReplica() error {
	mc := state.MasterClient
	if res, err := mc.Do([]string{"PING"}); err != nil {
		return err
	} else if !resp.Is(res, resp.RESPSimpleString{Value: "PONG"}) {
		return errors.New("expected PONG, got " + res.SerializeRESP())
	}
	port := config.Get("port")
	if res, err := mc.Do([]string{"REPLCONF", "listening-port", port}); err != nil {
		return err
	} else if !resp.Is(res, resp.RESPSimpleString{Value: "OK"}) {
		return errors.New("expected OK, got " + res.SerializeRESP())
	}
	if res, err := mc.Do([]string{"REPLCONF", "capa", "eof", "capa", "psync2"}); err != nil {
		return err
	} else if !resp.Is(res, resp.RESPSimpleString{Value: "OK"}) {
		return errors.New("expected OK, got " + res.SerializeRESP())
	}
	return nil
}

func performInitialSync() error {
	mc := state.MasterClient
	res, err := mc.Do([]string{"PSYNC", state.MasterReplid, strconv.FormatInt(state.MasterReplOffset.Load(), 10)})
	if err != nil {
		return err
	}
	rss, ok := res.(*resp.RESPSimpleString)
	if !ok {
		return errors.New("expected simple string, got " + res.SerializeRESP())
	}
	sa := strings.Split(rss.Value, " ")
	if !(len(sa) == 3 && sa[0] == "FULLRESYNC" && len(sa[1]) == 40 && sa[2] == "0") {
		return fmt.Errorf("invalid response: %s, expected only a FULLRESYNC command from a master with no history, in this dummy implementation", rss.Value)
	}
	state.MasterReplid = sa[1]
	MasterReplOffset, err := strconv.ParseInt(sa[2], 10, 64)
	if err != nil {
		return err
	}
	state.MasterReplOffset.Store(MasterReplOffset)

	// read RDB: dummy implementation; proper implementation uses a stream
	if b, err := mc.ReadByte(); err != nil {
		return err
	} else if b != '$' {
		return errors.New("expected '$', got " + string(b))
	}
	rdbr := respreader.NewBufBulkStringReader(mc.Reader)
	rdbresp, err := rdbr.ReadRESPUnterminated()
	if err != nil {
		log.Println("performHandshakeAsSlave: error reading RDB", err)
		return err
	}
	rdbstr, ok := rdbresp.(*resp.RESPBulkString)
	if !ok {
		return errors.New("expected bulk string, got " + rdbresp.SerializeRESP())
	}
	rdbreader.ReadRDBToState(bufio.NewReader(strings.NewReader(rdbstr.Value)), UnsafeResetDbWithSizeHint, UnsafeSet)
	log.Println("performHandshakeAsSlave: received RDB")
	return nil
}

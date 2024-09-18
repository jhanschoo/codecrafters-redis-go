package replication

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func InitializeSlave(replInfo *ReplicationInfo, replicaof string) error {
	log.Println("initializeSlave: started")
	replInfo.Role = "slave"

	var err error
	replInfo.masterClient, err = client.NewReplicaClient(replicaof)
	if err != nil {
		replInfo.masterClient.Close()
		return err
	}

	err = performHandshakeAsSlave(replInfo)
	// while incomplete, we close the connection
	replInfo.masterClient.Close()
	return err
}

func performHandshakeAsSlave(replInfo *ReplicationInfo) error {
	mc := replInfo.masterClient
	if res, err := mc.Do([]string{"PING"}); err != nil {
		return err
	} else if !resp.Is(res, resp.RESPSimpleString{Value: "PONG"}) {
		return errors.New("expected PONG, got " + res.SerializeRESP())
	}
	port, _ := config.Get("port")
	if res, err := mc.Do([]string{"REPLCONF", "listening-port", port}); err != nil {
		return err
	} else if !resp.Is(res, resp.RESPSimpleString{Value: "OK"}) {
		return errors.New("expected OK, got " + res.SerializeRESP())
	}
	if res, err := mc.Do([]string{"REPLCONF", "capa", "psync2"}); err != nil {
		return err
	} else if !resp.Is(res, resp.RESPSimpleString{Value: "OK"}) {
		return errors.New("expected OK, got " + res.SerializeRESP())
	}
	res, err := mc.Do([]string{"PSYNC", replInfo.MasterReplid, strconv.Itoa(replInfo.MasterReplOffset)})
	if err != nil {
		return err
	}
	rss, ok := res.(*resp.RESPSimpleString)
	if !ok {
		return errors.New("expected simple string, got " + res.SerializeRESP())
	}
	sa := strings.Split(rss.Value, " ")
	if !(len(sa) == 3 && sa[0] == "FULLRESYNC" && len(sa[1]) == 40 && sa[2] == "0") {
		return errors.New("invalid response: " + rss.Value)
	}
	return nil
}

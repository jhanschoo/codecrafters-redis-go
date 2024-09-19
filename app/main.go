package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/daemon"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

func main() {
	config.InitializeConfig()
	state.InitializeState()
	replication.InitializeReplication()

	daemon.LaunchDaemons()

	ri := replication.GetReplicationInfo()
	if ri.Role == "slave" {
		mc := ri.MasterClient
		go server.HandleConn(mc.Conn, mc.Reader, true)
	}
	// at the time of writing, this code below allows mutation on a replica
	server.Serve()
}

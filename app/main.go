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

	server.Serve()
}

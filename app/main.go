package main

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/daemon"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

func main() {
	config.InitializeConfig()
	state.Initialize()

	daemon.LaunchDaemons()

	if state.IsReplica() {
		log.Println("Replica beginning to listen for replication stream")
		go server.HandleReader(state.MasterClient().BufferedRESPConnReader)
	}

	server.Serve()
}

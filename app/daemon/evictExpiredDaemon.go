package daemon

import (
	"log"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/state"
)

func evictExpiredDaemon() {
	// TODO: implement channel to stop daemon
	const evictExpiredDaemonTimeBetweenRuns = 10 * time.Second
	log.Println("evictExpiredDaemon: started")
	for {
		time.Sleep(evictExpiredDaemonTimeBetweenRuns)
		// Note that this is a no-op for replicas
		state.SyncTryEvictExpiredKeysSweep()
	}
}

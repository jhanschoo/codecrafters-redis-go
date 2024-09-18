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
		state.SyncTryEvictExpiredKeysSweep()
	}
}

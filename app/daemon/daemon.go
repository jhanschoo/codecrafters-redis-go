package daemon

import "log"

var daemons = []func(){
	evictExpiredDaemon,
}

func Register(daemon func()) {
	if initialized {
		log.Panicln("Register: already initialized")
	}
	daemons = append(daemons, daemon)
}

var initialized = false

func LaunchDaemons() {
	if initialized {
		log.Panicln("InitializeDaemons: already initialized")
	}
	initialized = true
	for _, daemon := range daemons {
		go daemon()
	}
}

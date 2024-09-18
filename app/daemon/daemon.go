package daemon

var daemons = []func(){evictExpiredDaemon}

func InitializeDaemons() {
	for _, daemon := range daemons {
		go daemon()
	}
}

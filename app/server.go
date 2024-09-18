package main

import (
	"log"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/daemon"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

func main() {
	log.Println("Logs from your program will appear here!")

	config.InitializeConfig()
	port, ok := config.Get("port")
	if !ok {
		log.Println("Failed to read port from config")
		os.Exit(1)
	}

	state.InitializeState()

	daemon.InitializeDaemons()

	l, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Println("Failed to bind to port", port)
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(c)
	}

}

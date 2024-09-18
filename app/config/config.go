package config

import (
	"flag"
	"log"
)

var initialized = false

var config = map[string]*string{
	"dir":        flag.String("dir", ".", "directory to search for the RDB file"),
	"dbfilename": flag.String("dbfilename", "dump.rdb", "RDB file name"),
	"port":       flag.String("port", "6379", "port to listen on"),
	"replicaof":  flag.String("replicaof", "", "replicaof host port"),
}

func Get(key string) (string, bool) {
	s, a := config[key]
	return *s, a
}

func InitializeConfig() {
	if initialized {
		log.Fatalln("config: already initialized")
	}
	flag.Parse()
	initialized = true
}

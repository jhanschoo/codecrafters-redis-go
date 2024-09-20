package config

import (
	"flag"
	"log"
)

var initialized = false
var keys []string

var config = map[string]*string{
	"dir":        flag.String("dir", ".", "directory to search for the RDB file"),
	"dbfilename": flag.String("dbfilename", "dump.rdb", "RDB file name"),
	"port":       flag.String("port", "6379", "port to listen on"),
	"replicaof":  flag.String("replicaof", "", "replicaof host port"),
}

func Get(key string) string {
	if !initialized {
		log.Fatalln("config: not initialized")
	}
	s, a := config[key]
	if !a {
		log.Fatalf("config: key %s not found", key)
	}
	return *s
}

func Keys() []string {
	if !initialized {
		log.Fatalln("config: not initialized")
	}
	k := make([]string, len(keys))
	copy(k, keys)
	return k
}

func InitializeConfig() {
	if initialized {
		log.Fatalln("config: already initialized")
	}
	flag.Parse()
	initialized = true
	keys = make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
}

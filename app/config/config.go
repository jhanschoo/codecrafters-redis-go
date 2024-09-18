package config

import (
	"flag"
)

var config = map[string]*string{
	"dir":        flag.String("dir", ".", "directory to search for the RDB file"),
	"dbfilename": flag.String("dbfilename", "dump.rdb", "RDB file name"),
	"port":       flag.String("port", "6379", "port to listen on"),
}

func Get(key string) (string, bool) {
	s, a := config[key]
	return *s, a
}

func InitializeConfig() {
	flag.Parse()
}

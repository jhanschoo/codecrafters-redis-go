package config

import (
	"flag"
)

var config = map[string]*string{
	"dir":        flag.String("dir", ".", "directory to search for the RDB file"),
	"dbfilename": flag.String("dbfilename", "dump.rdb", "RDB file name"),
}

func Get(key string) (string, bool) {
	s, a := config[key]
	return *s, a
}

func ParseConfig() {
	flag.Parse()
}

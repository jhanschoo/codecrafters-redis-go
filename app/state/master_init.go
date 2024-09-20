package state

import (
	"bufio"
	"errors"
	"io/fs"
	"log"
	"os"
	"path"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/rdbreader"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

func initializeMaster() {
	initializeReplInfoAsMaster()
	initializeDbAsMaster()
}

func initializeReplInfoAsMaster() {
	state.Role = "master"
	state.MasterReplid = utility.RandomAlphaNumericString(40)
	// state.MasterReplOffset == 0
	// state.MasterClient == nil
}

func initializeDbAsMaster() {
	dir := config.Get("dir")
	dbfilename := config.Get("dbfilename")
	filePath := path.Join(dir, dbfilename)
	f, err := os.Open(path.Join(dir, dbfilename))
	if errors.Is(err, fs.ErrNotExist) {
		log.Printf("RDB file %s does not exist, skipping initialization from RDB file", filePath)
	} else {
		// defer is OK since we don'b care about handling the error here
		defer f.Close()
		if err != nil {
			log.Fatalf("failed to open RDB file: %v", err)
		}
		br := bufio.NewReader(f)
		rdbreader.ReadRDBToState(br, UnsafeResetDbWithSizeHint, UnsafeSet)
		return
	}
	log.Println("No RDB file specified or file does not exist, initializing empty db state")
	state.Db = make(map[string]DbValue)
}

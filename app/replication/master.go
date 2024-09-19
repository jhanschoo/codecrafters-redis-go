package replication

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

func initializeMaster() error {
	log.Println("initializeMaster: started")
	replicationInfo.Role = "master"
	replicationInfo.MasterReplid = utility.RandomAlphaNumericString(40)
	replicationInfo.MasterReplOffset = 0
	return nil
}

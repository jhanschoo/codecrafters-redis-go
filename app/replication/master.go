package replication

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

func initializeMaster(replInfo *ReplicationInfo) error {
	log.Println("initializeMaster: started")
	replInfo.Role = "master"
	replInfo.MasterReplid = utility.RandomAlphaNumericString(40)
	replInfo.MasterReplOffset = 0
	return nil
}

package replication

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

func initializeMaster(replInfo *ReplicationInfo) error {
	log.Println("initializeMaster: started")
	replicationInfo.Role = "master"
	replicationInfo.MasterReplid = utility.RandomAlphaNumericString(40)
	return nil
}

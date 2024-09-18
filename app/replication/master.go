package replication

import "github.com/codecrafters-io/redis-starter-go/app/utility"

func initializeMaster(replInfo *ReplicationInfo) error {
	replicationInfo.Role = "master"
	replicationInfo.MasterReplid = utility.RandomAlphaNumericString(40)
	return nil
}

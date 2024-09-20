package command

import (
	"errors"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var (
	respOk   = &resp.RESPSimpleString{Value: "OK"}
	respNull = &resp.RESPNull{CompatibilityFlag: 1}
)

// a standard subhandler responds with a RESP, that is the response to the client
// exactly when the connection is not the replica-master connection
type standardSubhandler func(sa []string, ctx Context) (resp.RESP, error)

type subhandler func(sa []string, ctx Context) error

var handlers = map[string]subhandler{
	pingCommand:     standard(handlePing),
	echoCommand:     standard(handleEcho),
	setCommand:      standard(handleSet),
	getCommand:      standard(handleGet),
	configCommand:   standard(handleConfigCommands),
	keysCommand:     standard(handleKeys),
	infoCommand:     standard(handleInfo),
	replconfCommand: handleReplconf,
	psyncCommand:    handlePsync,
	waitCommand:     standard(handleWait),
}

type Context struct {
	Reader     *respreader.BufferedRESPConnReader
	IsReplica  bool
	IsReplConn bool
	ReplOffset int64
	Com        resp.RESP
}

func standard(h standardSubhandler) subhandler {
	return func(sa []string, ctx Context) error {
		res, err := h(sa, ctx)
		if err != nil {
			return err
		}
		if ctx.IsReplConn {
			return nil
		}
		return writeRESP(ctx.Reader.Conn, res)
	}
}

func Handle(com resp.RESP, r *respreader.BufferedRESPConnReader) error {
	ctx := Context{
		Reader:     r,
		IsReplica:  state.IsReplica(),
		IsReplConn: state.IsReplConn(r),
		ReplOffset: state.ReplOffset(),
		Com:        com,
	}
	sa, ok := resp.DecodeStringSlice(com)
	if !ok || len(sa) == 0 {
		return errors.New("invalid input: expected non-empty array of bulk strings")
	}
	sh, ok := handlers[strings.ToUpper(sa[0])]
	if !ok {
		return writeRESPError(r.Conn, errors.New("unsupported command"))
	}
	err := sh(sa, ctx)
	if ctx.IsReplConn {
		// reserializing the command to determine bytes read
		// instead of tracking bytes read in the reader for convenience
		// this is a bit of a hack, but it's fine for this implementation
		//
		// in the replica, mutations to the replication offset all occur on the goroutine that handles the PSYNC command.
		// when the master sends a mutation to the replica, the offset increment is handled by state.ExecuteAndReplicateCommand for reasons of dataset-offset consistency.
		// otherwise, it is some form of liveness or sync command that the
		// master has sent to the replica, and exactly in this case the offset will not have changed (due to increments happening only on the PSYNC command handler)
		// we can then do a CAS operation to handle exactly this case
		state.CasOffset(ctx.ReplOffset, ctx.ReplOffset+int64(len(ctx.Com.SerializeRESP())))
	}
	return err
}

func HandleNext(r *respreader.BufferedRESPConnReader) error {
	com, err := r.ReadRESP()
	if err != nil {
		return err
	}
	log.Println("handleNext: received request", strconv.Quote(com.SerializeRESP()))
	return Handle(com, r)
}

func writeRESP(c net.Conn, res resp.RESP) error {
	_, err := c.Write([]byte(res.SerializeRESP()))
	return err
}

func writeRESPError(c net.Conn, err error) error {
	errMsg := err.Error()
	if strings.ContainsAny(errMsg, "\r\n") {
		return writeRESP(c, &resp.RESPBulkError{Value: errMsg})
	}
	return writeRESP(c, &resp.RESPSimpleError{Value: errMsg})
}

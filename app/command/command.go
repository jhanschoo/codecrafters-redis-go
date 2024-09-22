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

// a standard subhandler responds with a RESP, that is the response to the client
// exactly when the connection is not the replica-master connection
type subhandler func(sa []string, ctx Context) (resp.RESP, error)

var subhandlerMap = map[string]subhandler{
	pingCommand:     handlePing,
	echoCommand:     handleEcho,
	setCommand:      handleSet,
	getCommand:      handleGet,
	configCommand:   handleConfigCommands,
	keysCommand:     handleKeys,
	infoCommand:     handleInfo,
	waitCommand:     handleWait,
	typeCommand:     handleType,
	xaddCommand:     handleXadd,
	xrangeCommand:   handleXrange,
	xreadCommand:    handleXread,
	incrCommand:     handleIncr,
	multiCommand:    handleMulti,
	execCommand:     handleExec,
	psyncCommand:    handlePsync,
	replconfCommand: handleReplconf,
	discardCommand:  handleDiscard,
}

type Context struct {
	Reader        *respreader.BufferedRESPConnReader
	IsReplica     bool
	IsReplConn    bool
	ReplOffset    int64
	Com           resp.RESP
	Queued        *resp.ComSlice
	InTransaction bool
	Handle        func(Context) (resp.RESP, error)
}

type HandlerOptions struct {
	Queued        *resp.ComSlice
	InTransaction bool
}

func Handle(ctx Context) (resp.RESP, error) {
	com := ctx.Com
	log.Println("Handle: received request", strconv.Quote(com.SerializeRESP()), "isReplica:", ctx.IsReplica, "isReplConn:", ctx.IsReplConn, "queued:", ctx.Queued.IsActive())

	sa, ok := resp.DecodeStringSlice(com)
	if !ok || len(sa) == 0 {
		return nil, errors.New("invalid input: expected non-empty array of bulk strings")
	}
	if shouldQueueCommandInstead(sa, ctx) {
		ctx.Queued.AppendCom(ctx.Com)
		return resp.QueuedLit, nil
	}
	sh, ok := subhandlerMap[strings.ToUpper(sa[0])]
	if !ok {
		return &resp.RESPSimpleError{Value: `Unsupported command`}, nil
	}
	res, err := sh(sa, ctx)
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
	if err != nil {
		return nil, err
	}
	if shouldNotRespond(sa, ctx) {
		return nil, nil
	}
	return res, nil
}

func HandleNext(r *respreader.BufferedRESPConnReader, opts HandlerOptions) error {
	com, err := r.ReadRESP()
	if err != nil {
		return err
	}
	ctx := Context{
		Reader:        r,
		IsReplica:     state.IsReplica(),
		IsReplConn:    state.IsReplConn(r),
		ReplOffset:    state.ReplOffset(),
		Com:           com,
		Queued:        opts.Queued,
		InTransaction: opts.InTransaction,
		Handle:        Handle,
	}
	res, err := ctx.Handle(ctx)
	if err != nil {
		return err
	}
	if res != nil {
		return writeRESP(r.Conn, res)
	}
	return nil
}

func writeRESP(c net.Conn, res resp.RESP) error {
	_, err := c.Write([]byte(res.SerializeRESP()))
	return err
}

func shouldNotRespond(sa []string, ctx Context) bool {
	return ctx.IsReplConn && strings.ToUpper(sa[0]) != replconfCommand
}

func shouldQueueCommandInstead(sa []string, ctx Context) bool {
	if !ctx.Queued.IsActive() {
		return false
	}
	switch strings.ToUpper(sa[0]) {
	case multiCommand, execCommand, discardCommand:
		return false
	}
	return true
}

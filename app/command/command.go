package command

import (
	"errors"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func init() {
	defaultHandler.registerStandard(pingCommand, handlePing)
	defaultHandler.registerBasic(echoCommand, handleEcho)
	defaultHandler.registerStandard(setCommand, handleSet)
	defaultHandler.registerBasic(getCommand, handleGet)
	defaultHandler.registerBasic(configCommand, handleConfigCommands)
	defaultHandler.registerBasic(keysCommand, handleKeys)
	defaultHandler.registerBasic(infoCommand, handleInfo)
	defaultHandler.registerStandard(replconfCommand, handleReplconf)
	defaultHandler.register(psyncCommand, handlePsync)
}

type Context struct {
	Conn                    net.Conn
	Db                      int64
	IsReplica               bool
	IsPrivileged            bool
	BytesProcessed          int64
	ExecuteAndWriteToSlaves func(func() error, []string)
}

type Handler interface {
	Do(com resp.RESP, ctx Context) error
}

func (ch *CommandHandler) registerBasic(com string, do func(sa []string, db int64) (resp.RESP, error)) {
	ch.registerStandard(com, func(sa []string, ctx Context) (resp.RESP, error) {
		res, err := do(sa, ctx.Db)
		return res, err
	})
}

func (ch *CommandHandler) registerStandard(com string, do func(sa []string, ctx Context) (resp.RESP, error)) {
	ch.register(com, func(sa []string, ctx Context) error {
		res, err := do(sa, ctx)
		if err != nil {
			return err
		}
		if res == nil {
			return nil
		}
		_, err = ctx.Conn.Write([]byte(res.SerializeRESP()))
		return err
	})
}

type subhandler struct {
	Command string
	Do      func(sa []string, ctx Context) error
}

func (ch *CommandHandler) register(com string, do func(sa []string, ctx Context) error) {
	ch.handlers[com] = subhandler{Command: com, Do: do}
}

var defaultHandler = CommandHandler{
	handlers: make(map[string]subhandler),
}

func GetDefaultHandler() Handler {
	return &defaultHandler
}

type CommandHandler struct {
	handlers map[string]subhandler
}

func (h *CommandHandler) Do(com resp.RESP, ctx Context) error {
	sa, ok := resp.DecodeStringSlice(com)
	if !ok || len(sa) == 0 {
		return errors.New("invalid input: expected non-empty array of bulk strings")
	}
	sh, ok := h.handlers[strings.ToUpper(sa[0])]
	if !ok {
		return writeRESPError(ctx.Conn, errors.New("unsupported command"))
	}
	return sh.Do(sa, ctx)
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

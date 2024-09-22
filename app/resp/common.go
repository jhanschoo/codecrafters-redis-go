package resp

var (
	OkLit     = &RESPSimpleString{Value: "OK"}
	QueuedLit = &RESPSimpleString{Value: "QUEUED"}
	NullLit   = &RESPNull{CompatibilityFlag: 1}
)

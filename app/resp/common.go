package resp

var (
	OkLit   = &RESPSimpleString{Value: "OK"}
	NullLit = &RESPNull{CompatibilityFlag: 1}
)

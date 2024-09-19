package resp

import (
	"math/big"
	"strconv"
	"strings"
)

// Note: RESP types that refer to a sequence of bytes are implemented here as strings, since strings in Go are syntactic sugar for immutable byte arrays. This helps with ownership issues. Note however that aggregate types like RESPArray, RESPMap, etc. are implemented as slices of RESP, so the do not necessarily completely own the RESP objects they contain.

type RESP interface {
	SerializeRESP() string
}

type RESPSimpleString struct {
	Value string
}

var _ RESP = (*RESPSimpleString)(nil)

func (r RESPSimpleString) SerializeRESP() string {
	return "+" + r.Value + "\r\n"
}

type RESPSimpleError struct {
	Value string
}

var _ RESP = (*RESPSimpleError)(nil)

func (r RESPSimpleError) SerializeRESP() string {
	return "-" + r.Value + "\r\n"
}

type RESPInteger struct {
	Value int64
}

var _ RESP = (*RESPInteger)(nil)

func (r RESPInteger) SerializeRESP() string {
	return ":" + strconv.FormatInt(r.Value, 10) + "\r\n"
}

type RESPBulkString struct {
	Value string
}

var _ RESP = (*RESPBulkString)(nil)

func (r RESPBulkString) SerializeRESP() string {
	return "$" + strconv.Itoa(len(r.Value)) + "\r\n" + r.Value + "\r\n"
}

type RESPArray struct {
	Value []RESP
}

var _ RESP = (*RESPArray)(nil)

func (r RESPArray) SerializeRESP() string {
	var sb strings.Builder
	sb.WriteString("*" + strconv.Itoa(len(r.Value)) + "\r\n")
	for _, v := range r.Value {
		sb.WriteString(v.SerializeRESP())
	}
	return sb.String()
}

// compatibilityFlag is set to
// 1 if it serializes to "$-1\r\n"
// 2 if it serializes to "*-1\r\n"
// and any value (0 preferred) if it serializes to "_\r\n"
type RESPNull struct {
	CompatibilityFlag int
}

var _ RESP = (*RESPNull)(nil)

func (r RESPNull) SerializeRESP() string {
	switch r.CompatibilityFlag {
	case 1:
		return "$-1\r\n"
	case 2:
		return "*-1\r\n"
	default:
		return "_\r\n"
	}
}

type RESPBoolean struct {
	Value bool
}

var _ RESP = (*RESPBoolean)(nil)

func (r RESPBoolean) SerializeRESP() string {
	if r.Value {
		return "#t\r\n"
	}
	return "#f\r\n"
}

type RESPDouble struct {
	Value float64
}

var _ RESP = (*RESPDouble)(nil)

func (r RESPDouble) SerializeRESP() string {
	return "," + strings.ToLower(strconv.FormatFloat(r.Value, 'g', -1, 64)) + "\r\n"
}

type RESPBignum struct {
	Value big.Int
}

var _ RESP = (*RESPBignum)(nil)

func (r RESPBignum) SerializeRESP() string {
	return "(" + r.Value.String() + "\r\n"
}

type RESPBulkError struct {
	Value string
}

var _ RESP = (*RESPBulkError)(nil)

func (r RESPBulkError) SerializeRESP() string {
	return "!" + strconv.Itoa(len(r.Value)) + "\r\n" + r.Value + "\r\n"
}

type RESPVerbatimString struct {
	Encoding [3]byte
	Value    string
}

var _ RESP = (*RESPVerbatimString)(nil)

func (r RESPVerbatimString) SerializeRESP() string {
	return "=" + strconv.Itoa(len(r.Value)+4) + "\r\n" + string(r.Encoding[:]) + ":" + r.Value + "\r\n"
}

type RESPMap struct {
	Value []RESP
}

var _ RESP = (*RESPMap)(nil)

func (r RESPMap) SerializeRESP() string {
	var sb strings.Builder
	sb.WriteString("%" + strconv.Itoa(len(r.Value)/2) + "\r\n")
	for i := 0; i < len(r.Value); i += 2 {
		sb.WriteString(r.Value[i].SerializeRESP())
		sb.WriteString(r.Value[i+1].SerializeRESP())
	}
	return sb.String()
}

type RESPSet struct {
	Value []RESP
}

var _ RESP = (*RESPSet)(nil)

func (r RESPSet) SerializeRESP() string {
	var sb strings.Builder
	sb.WriteString("~" + strconv.Itoa(len(r.Value)) + "\r\n")
	for _, v := range r.Value {
		sb.WriteString(v.SerializeRESP())
	}
	return sb.String()
}

type RESPPush struct {
	Value []RESP
}

var _ RESP = (*RESPPush)(nil)

func (r RESPPush) SerializeRESP() string {
	var sb strings.Builder
	sb.WriteString(">" + strconv.Itoa(len(r.Value)) + "\r\n")
	for _, v := range r.Value {
		sb.WriteString(v.SerializeRESP())
	}
	return sb.String()
}

func EncodeStringSlice(sa []string) RESP {
	av := make([]RESP, len(sa))
	for i, s := range sa {
		av[i] = &RESPBulkString{Value: s}
	}
	return &RESPArray{Value: av}
}

func DecodeStringSlice(r RESP) ([]string, bool) {
	if ra, ok := r.(*RESPArray); ok {
		sa := make([]string, len(ra.Value))
		for i, v := range ra.Value {
			if bs, ok := v.(*RESPBulkString); ok {
				sa[i] = bs.Value
			} else {
				return nil, false
			}
		}
		return sa, true
	}
	return nil, false
}

func Is(r1 RESP, r2 RESP) bool {
	return r1.SerializeRESP() == r2.SerializeRESP()
}

package resp

import (
	"math/big"
	"strconv"
	"strings"
)

// Note: RESP types that refer to a sequence of bytes are implemented here as strings, since strings in Go are syntactic sugar for immutable byte arrays.

type RESP interface {
	SerializeRESP() string
}

type RESPSimpleString struct {
	Value string
}

func (r RESPSimpleString) SerializeRESP() string {
	return "+" + r.Value + "\r\n"
}

type RESPSimpleError struct {
	Value string
}

func (r RESPSimpleError) SerializeRESP() string {
	return "-" + r.Value + "\r\n"
}

type RESPInteger struct {
	Value int64
}

func (r RESPInteger) SerializeRESP() string {
	return ":" + strconv.FormatInt(r.Value, 10) + "\r\n"
}

type RESPBulkString struct {
	Value string
}

func (r RESPBulkString) SerializeRESP() string {
	return "$" + strconv.Itoa(len(r.Value)) + "\r\n" + r.Value + "\r\n"
}

type RESPArray struct {
	Value []RESP
}

func (r RESPArray) SerializeRESP() string {
	var sb strings.Builder
	sb.WriteString("*" + strconv.Itoa(len(r.Value)) + "\r\n")
	for _, v := range r.Value {
		sb.WriteString(v.SerializeRESP())
	}
	return sb.String()
}

type RESPNull struct{}

func (r RESPNull) SerializeRESP() string {
	return "_\r\n"
}

type RESPBoolean struct {
	Value bool
}

func (r RESPBoolean) SerializeRESP() string {
	if r.Value {
		return "#t\r\n"
	}
	return "#f\r\n"
}

type RESPDouble struct {
	Value float64
}

func (r RESPDouble) SerializeRESP() string {
	return "," + strings.ToLower(strconv.FormatFloat(r.Value, 'g', -1, 64)) + "\r\n"
}

type RESPBignum struct {
	Value big.Int
}

func (r RESPBignum) SerializeRESP() string {
	return "(" + r.Value.String() + "\r\n"
}

type RESPBulkError struct {
	Value string
}

func (r RESPBulkError) SerializeRESP() string {
	return "!" + strconv.Itoa(len(r.Value)) + "\r\n" + r.Value + "\r\n"
}

type RESPVerbatimString struct {
	Encoding [3]byte
	Value    string
}

func (r RESPVerbatimString) SerializeRESP() string {
	return "=" + strconv.Itoa(len(r.Value)+4) + "\r\n" + string(r.Encoding[:]) + ":" + r.Value + "\r\n"
}

type RESPMap struct {
	Value []RESP
}

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

func (r RESPPush) SerializeRESP() string {
	var sb strings.Builder
	sb.WriteString(">" + strconv.Itoa(len(r.Value)) + "\r\n")
	for _, v := range r.Value {
		sb.WriteString(v.SerializeRESP())
	}
	return sb.String()
}

package resp

import (
	"math/big"
	"strconv"
	"strings"
)

type RESP interface {
	Serialize() []byte
}

type RESPSimpleString struct {
	Value string
}

func (r RESPSimpleString) Serialize() []byte {
	return []byte("+" + r.Value + "\r\n")
}

type RESPError struct {
	Value string
}

func (r RESPError) Serialize() []byte {
	return []byte("-" + r.Value + "\r\n")
}

type RESPInteger struct {
	Value int64
}

func (r RESPInteger) Serialize() []byte {
	return []byte(":" + strconv.FormatInt(r.Value, 10) + "\r\n")
}

type RESPBulkString struct {
	Value []byte
}

func (r RESPBulkString) Serialize() (bs []byte) {
	bs = []byte("$" + strconv.Itoa(len(r.Value)) + "\r\n")
	bs = append(bs, r.Value...)
	bs = append(bs, []byte("\r\n")...)
	return
}

type RESPArray struct {
	Value []RESP
}

func (r RESPArray) Serialize() (bs []byte) {
	bs = []byte("*" + strconv.Itoa(len(r.Value)) + "\r\n")
	for _, v := range r.Value {
		bs = append(bs, v.Serialize()...)
	}
	return
}

type RESPNull struct{}

func (r RESPNull) Serialize() []byte {
	return []byte("_\r\n")
}

type RESPBoolean struct {
	Value bool
}

func (r RESPBoolean) Serialize() []byte {
	if r.Value {
		return []byte("#t\r\n")
	}
	return []byte("#f\r\n")
}

type RESPDouble struct {
	Value float64
}

func (r RESPDouble) Serialize() []byte {
	return []byte("," + strings.ToLower(strconv.FormatFloat(r.Value, 'g', -1, 64)) + "\r\n")
}

type RESPBignum struct {
	Value big.Int
}

func (r RESPBignum) Serialize() []byte {
	return []byte("(" + r.Value.String() + "\r\n")
}

type RESPBulkError struct {
	Value []byte
}

func (r RESPBulkError) Serialize() (bs []byte) {
	bs = []byte("!" + strconv.Itoa(len(r.Value)) + "\r\n")
	bs = append(bs, r.Value...)
	bs = append(bs, []byte("\r\n")...)
	return
}

type RESPVerbatimString struct {
	Encoding [3]byte
	Value    []byte
}

func (r RESPVerbatimString) Serialize() (bs []byte) {
	bs = []byte("=")
	bs = append(bs, r.Encoding[:]...)
	bs = append(bs, r.Value...)
	bs = append(bs, []byte("\r\n")...)
	return
}

type RESPMap struct {
	Value []RESP
}

func (r RESPMap) Serialize() (bs []byte) {
	bs = []byte("%" + strconv.Itoa(len(r.Value)/2) + "\r\n")
	for i := 0; i < len(r.Value); i += 2 {
		bs = append(bs, r.Value[i].Serialize()...)
		bs = append(bs, r.Value[i+1].Serialize()...)
	}
	return
}

type RESPSet struct {
	Value []RESP
}

func (r RESPSet) Serialize() (bs []byte) {
	bs = []byte("~" + strconv.Itoa(len(r.Value)) + "\r\n")
	for _, v := range r.Value {
		bs = append(bs, v.Serialize()...)
	}
	return
}

type RESPPush struct {
	Value []RESP
}

func (r RESPPush) Serialize() (bs []byte) {
	bs = []byte(">" + strconv.Itoa(len(r.Value)) + "\r\n")
	for _, v := range r.Value {
		bs = append(bs, v.Serialize()...)
	}
	return
}

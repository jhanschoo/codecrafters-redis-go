package state

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type DbStreamEntry struct {
	ms     int64
	seq    int64
	fields []string
}

func parseStreamEntryXaddId(id string) (int64, int64, error) {
	if id == "*" {
		return -1, -1, nil
	}
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0, ErrorInvalidIdFormat
	}
	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || ms < 0 {
		return 0, 0, ErrorInvalidIdFormat
	}
	if parts[1] == "*" {
		return ms, -1, nil
	}
	seq, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || seq < 0 {
		return 0, 0, ErrorInvalidIdFormat
	}
	return ms, seq, err
}

func parseStreamEntryXrangeId(id string, incr bool) (int64, int64, error) {
	if i := strings.IndexRune(id, '-'); i != -1 {
		ms, err := strconv.ParseInt(id[:i], 10, 64)
		if err != nil || ms < 0 {
			return 0, 0, ErrorInvalidIdFormat
		}
		seq, err := strconv.ParseInt(id[i+1:], 10, 64)
		if err != nil || seq < 0 {
			return 0, 0, ErrorInvalidIdFormat
		}
		if incr || (ms == 0 && seq == 0) {
			seq++
		}
		return ms, seq, nil
	}
	ms, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, 0, ErrorInvalidIdFormat
	}
	if incr {
		ms++
	}
	if ms == 0 {
		return 0, 1, nil
	}
	return ms, 0, nil
}

func parseStreamEntryXreadId(id string) (int64, int64, error) {
	var seqStr string = "0"
	if i := strings.IndexRune(id, '-'); i != -1 {
		seqStr = id[i+1:]
		id = id[:i]
	}
	ms, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, 0, ErrorInvalidIdFormat
	}
	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil || seq < 0 {
		return 0, 0, ErrorInvalidIdFormat
	}
	return ms, seq + 1, nil
}

func (e DbStreamEntry) Id() string {
	return strconv.FormatInt(e.ms, 10) + "-" + strconv.FormatInt(e.seq, 10)
}

type DbStream struct {
	data []DbStreamEntry
}

func (v *DbStream) Len() int {
	return len(v.data)
}

func (v *DbStream) Less(i, j int) bool {
	return v.data[i].ms < v.data[j].ms || (v.data[i].ms == v.data[j].ms && v.data[i].seq < v.data[j].seq)
}

func (v *DbStream) Swap(i, j int) {
	v.data[i], v.data[j] = v.data[j], v.data[i]
}

func (v *DbStream) nextValidId(ms, seq int64) (int64, int64, error) {
	if ms == 0 && seq == 0 {
		return 0, 0, ErrorInvalidIdValue
	}
	last := v.data[len(v.data)-1]
	if ms == -1 {
		ms = time.Now().UnixMilli()
		if last.ms > ms {
			ms = last.ms
		}
		seq = -1
	}
	if seq == -1 {
		if last.ms < ms {
			return ms, 0, nil
		}
		if last.ms == ms {
			return ms, last.seq + 1, nil
		}
		return 0, 0, ErrorInvalidNewId
	}
	if last.ms < ms || (last.ms == ms && last.seq < seq) {
		return ms, seq, nil
	}
	return 0, 0, ErrorInvalidNewId
}

func (v *DbStream) NewDbStreamEntryForStream(ms, seq int64, fields []string) (DbStreamEntry, error) {
	ms, seq, err := v.nextValidId(ms, seq)
	return DbStreamEntry{ms, seq, fields}, err
}

func (v *DbStream) SearchGreaterOrEqual(ms, seq int64) int {
	return sort.Search(len(v.data), func(i int) bool {
		return v.data[i].ms > ms || (v.data[i].ms == ms && v.data[i].seq >= seq)
	})
}

func (v *DbStream) EncodeSlice(i, j int) resp.RESP {
	if j == -1 {
		j = len(v.data)
	}
	slice := v.data[i:j]
	value := make([]resp.RESP, len(slice))
	for i, e := range slice {
		kv := make([]resp.RESP, 2)
		kv[0] = &resp.RESPBulkString{Value: e.Id()}
		v := make([]resp.RESP, len(e.fields))
		for i, f := range e.fields {
			v[i] = &resp.RESPBulkString{Value: f}
		}
		kv[1] = &resp.RESPArray{Value: v}
		value[i] = &resp.RESPArray{Value: kv}
	}
	return &resp.RESPArray{Value: value}
}

// Stream operations
var (
	ErrorInvalidIdFormat = errors.New("ERR Invalid stream ID format")
	ErrorInvalidIdValue  = errors.New("ERR The ID specified in XADD must be greater than 0-0")
	ErrorInvalidNewId    = errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
)

func Xadd(key string, id string, fields []string) (string, error) {
	ms, seq, err := parseStreamEntryXaddId(id)
	state.DbMu.Lock()
	defer state.DbMu.Unlock()
	v, ok := state.Db[key]
	if !ok {
		v = &DbStream{data: []DbStreamEntry{
			{ms: 0, seq: 0, fields: nil},
		}}
		state.Db[key] = v
	}
	stream, ok := v.(*DbStream)
	if !ok {
		return "", ErrorWrongType
	}
	e, err := stream.NewDbStreamEntryForStream(ms, seq, fields)
	if err != nil {
		return "", err
	}
	stream.data = append(stream.data, e)
	return e.Id(), nil
}

func Xrange(key, start, end string) (resp.RESP, error) {
	var startMs, startSeq, endMs, endSeq int64
	var startIndex, endIndex int
	var err error
	if start != "-" {
		startMs, startSeq, err = parseStreamEntryXrangeId(start, false)
	}
	if err != nil {
		return nil, err
	}
	if end != "+" {
		endMs, endSeq, err = parseStreamEntryXrangeId(end, true)
	}
	if err != nil {
		return nil, err
	}
	state.DbMu.RLock()
	v, ok := state.Db[key]
	if !ok {
		return nil, ErrorNone
	}
	stream, ok := v.(*DbStream)
	if !ok {
		return nil, ErrorWrongType
	}
	if start != "-" {
		startIndex = stream.SearchGreaterOrEqual(startMs, startSeq)
	} else {
		startIndex = 1
	}
	if end != "+" {
		endIndex = stream.SearchGreaterOrEqual(endMs, endSeq)
	} else {
		endIndex = len(stream.data)
	}
	state.DbMu.RUnlock()
	return stream.EncodeSlice(startIndex, endIndex), nil
}

func Xread(key, start string) (resp.RESP, error) {
	startMs, startSeq, err := parseStreamEntryXreadId(start)
	if err != nil {
		return nil, err
	}
	state.DbMu.RLock()
	v, ok := state.Db[key]
	if !ok {
		return nil, ErrorNone
	}
	stream, ok := v.(*DbStream)
	if !ok {
		return nil, ErrorWrongType
	}
	startIndex := stream.SearchGreaterOrEqual(startMs, startSeq)
	state.DbMu.RUnlock()
	ss := stream.EncodeSlice(startIndex, -1)
	skv := make([]resp.RESP, 2)
	skv[0] = &resp.RESPBulkString{Value: key}
	skv[1] = ss
	skvr := &resp.RESPArray{Value: skv}
	skvs := make([]resp.RESP, 1)
	skvs[0] = skvr
	skvsa := &resp.RESPArray{Value: skvs}
	return skvsa, nil
}

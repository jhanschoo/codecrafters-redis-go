package state

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
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
	if id == "$" {
		return -1, -1, nil
	}
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
	ErrorInvalidInput    = errors.New("ERR Invalid input")
)

func Xadd(key string, id string, fields []string) (string, error) {
	ms, seq, err := parseStreamEntryXaddId(id)
	if err != nil {
		return "", err
	}
	state.DbMu.Lock()
	v, ok := state.Db[key]
	if !ok {
		v = &DbStream{data: []DbStreamEntry{
			{ms: 0, seq: 0, fields: nil},
		}}
		state.Db[key] = v
	}
	stream, ok := v.(*DbStream)
	if !ok {
		state.DbMu.Unlock()
		return "", ErrorWrongType
	}
	e, err := stream.NewDbStreamEntryForStream(ms, seq, fields)
	if err != nil {
		state.DbMu.Unlock()
		return "", err
	}
	stream.data = append(stream.data, e)
	ss := stream.EncodeSlice(len(stream.data)-1, len(stream.data))
	streamBlockListenersMu.Lock()
	skv := make([]resp.RESP, 2)
	skv[0] = &resp.RESPBulkString{Value: key}
	skv[1] = ss
	skvr := &resp.RESPArray{Value: skv}
	sssa := &resp.RESPArray{Value: []resp.RESP{skvr}}
	state.DbMu.Unlock()
	listeners, ok := streamBlockListeners[key]
	if ok {
		for listener := range listeners {
			listener.l.Lock()
			if listener.res == nullListenerRes {
				listener.res = sssa
			}
			listener.cond.Broadcast()
			listener.l.Unlock()
		}
	}
	streamBlockListenersMu.Unlock()
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
	defer state.DbMu.RUnlock()
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
	return stream.EncodeSlice(startIndex, endIndex), nil
}

func Xread(kids []string, blockTimeout time.Duration) (resp.RESP, error) {
	if len(kids)%2 != 0 {
		return nil, ErrorInvalidInput
	}
	n := len(kids) / 2
	keys := make([]string, n)
	mss := make([]int64, n)
	seqs := make([]int64, n)
	var err error
	for i := 0; i < n; i++ {
		mss[i], seqs[i], err = parseStreamEntryXreadId(kids[n+i])
		if err != nil {
			return nil, err
		}
		keys[i] = kids[i]
	}
	sss := make([]resp.RESP, 0, len(keys))
	state.DbMu.RLock()
	for i, key := range keys {
		v, ok := state.Db[key]
		if !ok {
			return nil, ErrorNone
		}
		stream, ok := v.(*DbStream)
		if !ok {
			return nil, ErrorWrongType
		}
		startIndex := stream.Len()
		if mss[i] != -1 {
			startIndex = stream.SearchGreaterOrEqual(mss[i], seqs[i])
		}

		// handle no new items
		if startIndex >= stream.Len() {
			continue
		}
		ss := stream.EncodeSlice(startIndex, -1)
		skv := make([]resp.RESP, 2)
		skv[0] = &resp.RESPBulkString{Value: key}
		skv[1] = ss
		skvr := &resp.RESPArray{Value: skv}
		sss = append(sss, skvr)
	}
	if len(sss) != 0 || blockTimeout < 0 {
		sssa := &resp.RESPArray{Value: sss}
		state.DbMu.RUnlock()
		return sssa, nil
	}
	// block
	listener := newStreamBlockListener(keys)
	listener.l.Lock()
	streamBlockListenersMu.Lock()
	for _, key := range keys {
		if streamBlockListeners[key] == nil {
			streamBlockListeners[key] = make(map[*streamBlockListener]bool, 1)
		}
		streamBlockListeners[key][listener] = true
	}
	state.DbMu.RUnlock()
	streamBlockListenersMu.Unlock()
	go utility.Timeout(blockTimeout, listener.l, listener.cond, nil)
	listener.cond.Wait()
	res := listener.res
	listener.l.Unlock()
	streamBlockListenersMu.Lock()
	for _, key := range keys {
		delete(streamBlockListeners[key], listener)
	}
	streamBlockListenersMu.Unlock()
	return res, nil
}

type streamBlockListener struct {
	l    *sync.Mutex
	cond *sync.Cond
	keys []string
	res  resp.RESP
}

func newStreamBlockListener(keys []string) *streamBlockListener {
	l := &sync.Mutex{}
	return &streamBlockListener{
		l:    l,
		cond: sync.NewCond(l),
		keys: keys,
		res:  nullListenerRes,
	}
}

var streamBlockListeners = make(map[string]map[*streamBlockListener]bool)

var streamBlockListenersMu sync.Mutex

var nullListenerRes = &resp.RESPNull{CompatibilityFlag: 1}

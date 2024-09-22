package utility

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

func RandomAlphaNumericString(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Timeout(t time.Duration, mu *sync.Mutex, cond *sync.Cond, f func() bool) {
	// if t == 0, the timeout is infinite
	if t == 0 {
		return
	}
	time.Sleep(t)
	mu.Lock()
	defer mu.Unlock()
	if f != nil && !f() {
		return
	}
	cond.Broadcast()
}

// INFO command serialization

type InfoValue interface {
	SerializeTo(sb *strings.Builder)
}

type InfoString string

func (i InfoString) SerializeTo(sb *strings.Builder) {
	sb.WriteString(string(i))
}

type InfoMap map[string]string

func (i InfoMap) SerializeTo(sb *strings.Builder) {
	var followerEntry bool
	for k, v := range i {
		if followerEntry {
			sb.WriteString(",")
		}
		followerEntry = true
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}
}

type Info map[string]InfoValue

func (i Info) Serialize() string {
	var sb strings.Builder
	for k, v := range i {
		sb.WriteString(k)
		sb.WriteString(":")
		v.SerializeTo(&sb)
		sb.WriteString("\r\n")
	}
	return sb.String()
}

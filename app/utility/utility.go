package utility

import (
	"math/rand"
	"strings"
)

func RandomAlphaNumericString(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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

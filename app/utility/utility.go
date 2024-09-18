package utility

import (
	"log"
	"math/rand"
	"strconv"
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

func SerializeInfo(info map[string]interface{}) string {
	var sb strings.Builder
	for k, w := range info {
		sb.WriteString(k)
		sb.WriteString(":")
		switch v := w.(type) {
		case string:
			sb.WriteString(v)
		case float64:
			sb.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		case map[string]interface{}:
			var followerEntry bool
			for k1, v1 := range v {
				if followerEntry {
					sb.WriteString(",")
				}
				v1Str, ok := v1.(string)
				if !ok {
					log.Fatalf("unexpected type %T while serializing info", v1)
				}
				followerEntry = true
				sb.WriteString(k1)
				sb.WriteString("=")
				sb.WriteString(v1Str)
			}
		default:
			log.Fatalf("unexpected type %T while serializing info", v)
		}
		sb.WriteString("\r\n")
	}
	return sb.String()
}

package command

import "sync"

type kvValue struct {
	value     string
	expiresAt int64
}

var stateRWMutex = sync.RWMutex{}

var state = make(map[string]kvValue)

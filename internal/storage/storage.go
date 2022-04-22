package storage

import "sync"

var Counters = make(map[string]int64)
var Gauges = make(map[string]float64)
var CounterMx sync.Mutex
var GaugeMx sync.Mutex

package poller

import (
	"context"
	"sync"
	"testing"
	"time"
)

func BenchmarkGetRtmMetrics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getRtmMetrics()
	}
}

func BenchmarkGetMemoryMetrics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getMemoryMetrics()
	}
}

func BenchmarkGetCPUMetrics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getCPUMetrics()
	}
}

func TestCUint64(t *testing.T) {
	name := "testing cUint64"
	var value uint64 = 1
	t.Run(name, func(t *testing.T) {
		cUint64(value)
	})

}

func TestGetRtmMetrics(t *testing.T) {
	name := "testing getRtmMetrics"
	t.Run(name, func(t *testing.T) {
		getRtmMetrics()
	})
}

func TestGetMemoryMetrics(t *testing.T) {
	name := "testing getMemoryCPUMetrics"
	t.Run(name, func(t *testing.T) {
		_, err := getMemoryMetrics()
		if err != nil {
			t.Errorf("getMemoryMetrics() has returned an error: %s", err)
		}
	})
}

func TestGetCPUMetrics(t *testing.T) {
	name := "testing getCPUMetrics"
	t.Run(name, func(t *testing.T) {
		_, err := getCPUMetrics()
		if err != nil {
			t.Errorf("getCPUMetrics() has returned an error: %s", err)
		}
	})
}

func TestPoll(t *testing.T) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)

	name := "testing Pool"
	t.Run(name, func(t *testing.T) {
		Poll(ctxTimeout, &wg, time.Second)
	})

}

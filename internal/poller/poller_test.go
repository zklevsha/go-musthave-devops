package poller

import (
	"testing"
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

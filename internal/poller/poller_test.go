package poller

import (
	"testing"
)

func BenchmarkGetRtmMetrics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getRtmMetrics()
	}
}

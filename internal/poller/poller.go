package poller

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/zklevsha/go-musthave-devops/internal/storage"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

func cUint64(i uint64) *float64 {
	res := float64(i)
	return &res
}

func Poll(ctx context.Context, wg *sync.WaitGroup, pollInterval time.Duration) {
	defer wg.Done()
	ticker := time.NewTicker(pollInterval)
	for {
		select {
		case <-ctx.Done():
			log.Println("INFO poll received ctx.Done(), returning")
			return
		case <-ticker.C:
			log.Println("INFO polling data")
			var rtm runtime.MemStats
			runtime.ReadMemStats(&rtm)
			random := rand.Float64()
			pollCount := int64(1)
			var metrics = []structs.Metric{
				{MType: "gauge", ID: "Alloc", Value: cUint64(rtm.Alloc)},
				{MType: "gauge", ID: "BuckHashSys", Value: cUint64(rtm.BuckHashSys)},
				{MType: "gauge", ID: "Frees", Value: cUint64(rtm.Frees)},
				{MType: "gauge", ID: "GCCPUFraction", Value: &rtm.GCCPUFraction},
				{MType: "gauge", ID: "GCSys", Value: cUint64(rtm.GCSys)},
				{MType: "gauge", ID: "HeapAlloc", Value: cUint64(rtm.HeapAlloc)},
				{MType: "gauge", ID: "HeapIdle", Value: cUint64(rtm.HeapIdle)},
				{MType: "gauge", ID: "HeapInuse", Value: cUint64(rtm.HeapInuse)},
				{MType: "gauge", ID: "HeapObjects", Value: cUint64(rtm.HeapObjects)},
				{MType: "gauge", ID: "HeapReleased", Value: cUint64(rtm.HeapReleased)},
				{MType: "gauge", ID: "HeapSys", Value: cUint64(rtm.HeapSys)},
				{MType: "gauge", ID: "LastGC", Value: cUint64(rtm.LastGC)},
				{MType: "gauge", ID: "Lookups", Value: cUint64(rtm.Lookups)},
				{MType: "gauge", ID: "MCacheInuse", Value: cUint64(rtm.MCacheInuse)},
				{MType: "gauge", ID: "MCacheSys", Value: cUint64(rtm.MCacheSys)},
				{MType: "gauge", ID: "MSpanSys", Value: cUint64(rtm.MSpanSys)},
				{MType: "gauge", ID: "Mallocs", Value: cUint64(rtm.Mallocs)},
				{MType: "gauge", ID: "NumForcedGC", Value: cUint64(uint64(rtm.NumForcedGC))},
				{MType: "gauge", ID: "NextGC", Value: cUint64(rtm.NextGC)},
				{MType: "gauge", ID: "OtherSys", Value: cUint64(rtm.OtherSys)},
				{MType: "gauge", ID: "PauseTotalNs", Value: cUint64(rtm.PauseTotalNs)},
				{MType: "gauge", ID: "StackInuse", Value: cUint64(rtm.StackInuse)},
				{MType: "gauge", ID: "StackSys", Value: cUint64(rtm.StackSys)},
				{MType: "gauge", ID: "Sys", Value: cUint64(rtm.Sys)},
				{MType: "gauge", ID: "TotalAlloc", Value: cUint64(rtm.TotalAlloc)},
				{MType: "gauge", ID: "RandomValue", Value: &random},
				{MType: "gauge", ID: "MSpanInuse", Value: cUint64(rtm.MSpanInuse)},
				{MType: "gauge", ID: "NumGC", Value: cUint64(uint64(rtm.NumGC))},
				{MType: "counter", ID: "PollCount", Delta: &pollCount},
			}
			err := storage.Agent.UpdateMetrics(metrics)
			if err != nil {
				log.Printf("ERROR poller failed to poll metrics: %s", err.Error())
			}
		}
	}
}

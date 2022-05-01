package poller

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

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
			storage.Agent.SetGauge("Alloc", float64(rtm.Alloc))
			storage.Agent.SetGauge("BuckHashSys", float64(rtm.BuckHashSys))
			storage.Agent.SetGauge("Frees", float64(rtm.Frees))
			storage.Agent.SetGauge("GCCPUFraction", float64(rtm.GCCPUFraction))
			storage.Agent.SetGauge("GCSys", float64(rtm.GCSys))
			storage.Agent.SetGauge("HeapAlloc", float64(rtm.HeapAlloc))
			storage.Agent.SetGauge("HeapIdle", float64(rtm.HeapIdle))
			storage.Agent.SetGauge("HeapInuse", float64(rtm.HeapInuse))
			storage.Agent.SetGauge("HeapObjects", float64(rtm.HeapObjects))
			storage.Agent.SetGauge("HeapReleased", float64(rtm.HeapReleased))
			storage.Agent.SetGauge("HeapSys", float64(rtm.HeapSys))
			storage.Agent.SetGauge("LastGC", float64(rtm.LastGC))
			storage.Agent.SetGauge("Lookups", float64(rtm.Lookups))
			storage.Agent.SetGauge("MCacheInuse", float64(rtm.MCacheInuse))
			storage.Agent.SetGauge("MCacheSys", float64(rtm.MCacheSys))
			storage.Agent.SetGauge("MSpanSys", float64(rtm.MSpanSys))
			storage.Agent.SetGauge("Mallocs", float64(rtm.Mallocs))
			storage.Agent.SetGauge("NextGC", float64(rtm.NextGC))
			storage.Agent.SetGauge("NumForcedGC", float64(rtm.NumForcedGC))
			storage.Agent.SetGauge("NextGC", float64(rtm.NumGC))
			storage.Agent.SetGauge("OtherSys", float64(rtm.OtherSys))
			storage.Agent.SetGauge("PauseTotalNs", float64(rtm.PauseTotalNs))
			storage.Agent.SetGauge("StackInuse", float64(rtm.StackInuse))
			storage.Agent.SetGauge("StackSys", float64(rtm.StackSys))
			storage.Agent.SetGauge("Sys", float64(rtm.Sys))
			storage.Agent.SetGauge("TotalAlloc", float64(rtm.TotalAlloc))
			storage.Agent.SetGauge("RandomValue", rand.Float64())
			storage.Agent.SetGauge("MSpanInuse", float64(rtm.MSpanInuse))
			storage.Agent.SetGauge("NumGC", float64(rtm.NumGC))
			storage.Agent.IncreaseCounter("PollCount", 1)
		}
	}
}

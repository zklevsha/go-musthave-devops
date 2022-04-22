package poller

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/zklevsha/go-musthave-devops/agent/storage"
)

func Poll(ctx context.Context, wg *sync.WaitGroup, pollInterval time.Duration) {
	var counter int64
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("INFO poll received ctx.Done(), returning")
			return
		default:
			log.Println("INFO polling data")
			counter++
			var rtm runtime.MemStats
			runtime.ReadMemStats(&rtm)
			storage.Mutex.Lock()
			storage.Gauges["Alloc"] = float64(rtm.Alloc)
			storage.Gauges["BuckHashSys"] = float64(rtm.BuckHashSys)
			storage.Gauges["Frees"] = float64(rtm.Frees)
			storage.Gauges["GCCPUFraction"] = float64(rtm.GCCPUFraction)
			storage.Gauges["GCSys"] = float64(rtm.GCSys)
			storage.Gauges["HeapAlloc"] = float64(rtm.HeapAlloc)
			storage.Gauges["HeapIdle"] = float64(rtm.HeapIdle)
			storage.Gauges["HeapInuse"] = float64(rtm.HeapInuse)
			storage.Gauges["HeapObjects"] = float64(rtm.HeapObjects)
			storage.Gauges["HeapReleased"] = float64(rtm.HeapReleased)
			storage.Gauges["HeapSys"] = float64(rtm.HeapSys)
			storage.Gauges["LastGC"] = float64(rtm.LastGC)
			storage.Gauges["Lookups"] = float64(rtm.Lookups)
			storage.Gauges["MCacheInuse"] = float64(rtm.MCacheInuse)
			storage.Gauges["MCacheSys"] = float64(rtm.MCacheSys)
			storage.Gauges["MSpanSys"] = float64(rtm.MSpanSys)
			storage.Gauges["Mallocs"] = float64(rtm.Mallocs)
			storage.Gauges["NextGC"] = float64(rtm.NextGC)
			storage.Gauges["NumForcedGC"] = float64(rtm.NumForcedGC)
			storage.Gauges["NextGC"] = float64(rtm.NumGC)
			storage.Gauges["OtherSys"] = float64(rtm.OtherSys)
			storage.Gauges["PauseTotalNs"] = float64(rtm.PauseTotalNs)
			storage.Gauges["StackInuse"] = float64(rtm.StackInuse)
			storage.Gauges["StackSys"] = float64(rtm.StackSys)
			storage.Gauges["Sys"] = float64(rtm.Sys)
			storage.Gauges["TotalAlloc"] = float64(rtm.TotalAlloc)
			storage.Gauges["RandomValue"] = rand.Float64()
			storage.Counters["PollCount"] = counter
			storage.Mutex.Unlock()
			time.Sleep(pollInterval)
		}
	}
}

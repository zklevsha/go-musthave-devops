package poller

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/zklevsha/go-musthave-devops/internal/agstore"
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
			agstore.Mutex.Lock()
			agstore.Gauges["Alloc"] = float64(rtm.Alloc)
			agstore.Gauges["BuckHashSys"] = float64(rtm.BuckHashSys)
			agstore.Gauges["Frees"] = float64(rtm.Frees)
			agstore.Gauges["GCCPUFraction"] = float64(rtm.GCCPUFraction)
			agstore.Gauges["GCSys"] = float64(rtm.GCSys)
			agstore.Gauges["HeapAlloc"] = float64(rtm.HeapAlloc)
			agstore.Gauges["HeapIdle"] = float64(rtm.HeapIdle)
			agstore.Gauges["HeapInuse"] = float64(rtm.HeapInuse)
			agstore.Gauges["HeapObjects"] = float64(rtm.HeapObjects)
			agstore.Gauges["HeapReleased"] = float64(rtm.HeapReleased)
			agstore.Gauges["HeapSys"] = float64(rtm.HeapSys)
			agstore.Gauges["LastGC"] = float64(rtm.LastGC)
			agstore.Gauges["Lookups"] = float64(rtm.Lookups)
			agstore.Gauges["MCacheInuse"] = float64(rtm.MCacheInuse)
			agstore.Gauges["MCacheSys"] = float64(rtm.MCacheSys)
			agstore.Gauges["MSpanSys"] = float64(rtm.MSpanSys)
			agstore.Gauges["Mallocs"] = float64(rtm.Mallocs)
			agstore.Gauges["NextGC"] = float64(rtm.NextGC)
			agstore.Gauges["NumForcedGC"] = float64(rtm.NumForcedGC)
			agstore.Gauges["NextGC"] = float64(rtm.NumGC)
			agstore.Gauges["OtherSys"] = float64(rtm.OtherSys)
			agstore.Gauges["PauseTotalNs"] = float64(rtm.PauseTotalNs)
			agstore.Gauges["StackInuse"] = float64(rtm.StackInuse)
			agstore.Gauges["StackSys"] = float64(rtm.StackSys)
			agstore.Gauges["Sys"] = float64(rtm.Sys)
			agstore.Gauges["TotalAlloc"] = float64(rtm.TotalAlloc)
			agstore.Gauges["RandomValue"] = rand.Float64()
			agstore.Counters["PollCount"] = counter
			agstore.Mutex.Unlock()
			time.Sleep(pollInterval)
		}
	}
}

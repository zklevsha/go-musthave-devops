package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const pollInterval = time.Duration(2 * time.Second)
const reportInterval = time.Duration(10 * time.Second)
const serverSocket = "localhost:8080"

var counters = map[string]int64{
	"PollCount": 0,
}
var gauges = map[string]float64{
	"Alloc":         0,
	"BuckHashSys":   0,
	"Frees":         0,
	"GCCPUFraction": 0,
	"GCSys":         0,
	"HeapAlloc":     0,
	"HeapIdle":      0,
	"HeapInuse":     0,
	"HeapObjects":   0,
	"HeapReleased":  0,
	"HeapSys":       0,
	"LastGC":        0,
	"Lookups":       0,
	"MCacheInuse":   0,
	"MCacheSys":     0,
	"MSpanInuse":    0,
	"MSpanSys":      0,
	"Mallocs":       0,
	"NextGC":        0,
	"NumForcedGC":   0,
	"NumGC":         0,
	"OtherSys":      0,
	"PauseTotalNs":  0,
	"StackInuse":    0,
	"StackSys":      0,
	"Sys":           0,
	"TotalAlloc":    0,
	"RandomValue":   0,
}

var mutex sync.Mutex

func poll() {
	var counter int64
	for {
		counter++
		var rtm runtime.MemStats
		runtime.ReadMemStats(&rtm)
		mutex.Lock()
		gauges["Alloc"] = float64(rtm.Alloc)
		gauges["BuckHashSys"] = float64(rtm.BuckHashSys)
		gauges["Frees"] = float64(rtm.Frees)
		gauges["GCCPUFraction"] = float64(rtm.GCCPUFraction)
		gauges["GCSys"] = float64(rtm.GCSys)
		gauges["HeapAlloc"] = float64(rtm.HeapAlloc)
		gauges["HeapIdle"] = float64(rtm.HeapIdle)
		gauges["HeapInuse"] = float64(rtm.HeapInuse)
		gauges["HeapObjects"] = float64(rtm.HeapObjects)
		gauges["HeapReleased"] = float64(rtm.HeapReleased)
		gauges["HeapSys"] = float64(rtm.HeapSys)
		gauges["LastGC"] = float64(rtm.LastGC)
		gauges["Lookups"] = float64(rtm.Lookups)
		gauges["MCacheInuse"] = float64(rtm.MCacheInuse)
		gauges["MCacheSys"] = float64(rtm.MCacheSys)
		gauges["MSpanSys"] = float64(rtm.MSpanSys)
		gauges["Mallocs"] = float64(rtm.Mallocs)
		gauges["NextGC"] = float64(rtm.NextGC)
		gauges["NumForcedGC"] = float64(rtm.NumForcedGC)
		gauges["NextGC"] = float64(rtm.NumGC)
		gauges["OtherSys"] = float64(rtm.OtherSys)
		gauges["PauseTotalNs"] = float64(rtm.PauseTotalNs)
		gauges["StackInuse"] = float64(rtm.StackInuse)
		gauges["StackSys"] = float64(rtm.StackSys)
		gauges["Sys"] = float64(rtm.Sys)
		gauges["TotalAlloc"] = float64(rtm.TotalAlloc)
		gauges["RandomValue"] = float64(rand.Int())
		counters["PollCount"] = counter
		mutex.Unlock()
		time.Sleep(pollInterval)
	}
}

func send(url string) {
	_, err := http.Post(url, "text/plain", bytes.NewBufferString(""))
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
}

func report() {
	time.Sleep(reportInterval)
	for {
		mutex.Lock()
		g := gauges
		c := counters
		mutex.Unlock()
		for k, v := range g {
			send(fmt.Sprintf("http://%s/update/%s/%s/%f", serverSocket, "gauge", k, v))
		}
		for k, v := range c {
			send(fmt.Sprintf("http://%s/update/%s/%s/%d", serverSocket, "counter", k, v))
		}
		time.Sleep(reportInterval)
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go poll()
	go report()
	wg.Wait()

}

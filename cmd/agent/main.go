package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const pollInterval = time.Duration(2 * time.Second)
const reportInterval = time.Duration(10 * time.Second)
const serverSocket = "127.0.0.1:8080"

var wg sync.WaitGroup

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

func poll(ctx context.Context) {
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
			gauges["RandomValue"] = rand.Float64()
			counters["PollCount"] = counter
			mutex.Unlock()
			time.Sleep(pollInterval)
		}
	}
}

func send(url string) error {
	resp, err := http.Post(url, "text/plain", bytes.NewBufferString(""))
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		return fmt.Errorf("bad StatusCode: %s (URL: %s, Response Body: %s)",
			url, resp.Status, string(body))
	}
	return nil
}

func reportGauges() {
	mutex.Lock()
	g := gauges
	mutex.Unlock()
	for k, v := range g {
		err := send(fmt.Sprintf("http://%s/update/%s/%s/%f", serverSocket, "gauge", k, v))
		if err != nil {
			log.Printf("ERROR failed to send metic %s(%f): %s\n", k, v, err.Error())
		} else {
			log.Printf("INFO metric %s(%f) was sent\n", k, v)
		}

	}
}

func reportCounters() {
	mutex.Lock()
	c := counters
	mutex.Unlock()

	for k, v := range c {
		err := send(fmt.Sprintf("http://%s/update/%s/%s/%d", serverSocket, "counter", k, v))
		if err != nil {
			log.Printf("ERROR failed to send metic %s(%d): %s\n", k, v, err.Error())
		} else {
			log.Printf("INFO metric %s(%d) was sent\n", k, v)
		}
	}
}

func report(ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("INFO report received ctx.Done(), returning")
			return
		default:
			reportGauges()
			reportCounters()
			time.Sleep(reportInterval)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	wg.Add(2)
	go poll(ctx)
	go report(ctx)
	sig := <-c
	log.Printf("INFO main got a signal '%v', start shutting down...\n", sig)
	cancel()
	wg.Wait()
	log.Printf("INFO main Shutdown complete")
}

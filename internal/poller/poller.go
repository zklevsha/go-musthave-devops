package poller

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

func cUint64(i uint64) *float64 {
	res := float64(i)
	return &res
}

func getRtmMetrics() []structs.Metric {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
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
		{MType: "gauge", ID: "MSpanInuse", Value: cUint64(rtm.MSpanInuse)},
		{MType: "gauge", ID: "NumGC", Value: cUint64(uint64(rtm.NumGC))},
	}
	return metrics
}

func getMemoryMetrics() ([]structs.Metric, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return []structs.Metric{}, err
	}
	metrics := []structs.Metric{
		{MType: "gauge", ID: "TotalMemory", Value: cUint64(v.Total)},
		{MType: "gauge", ID: "FreeMemory", Value: cUint64(v.Free)},
	}
	return metrics, nil

}

func getCPUMetrics() ([]structs.Metric, error) {
	var metrics = []structs.Metric{}
	cpu, err := cpu.Percent(0, true)
	if err != nil {
		return metrics, err
	}
	for i, c := range cpu {
		m := structs.Metric{
			MType: "gauge",
			ID:    fmt.Sprintf("CPUutilization%d", i+1),
			Value: &c}
		metrics = append(metrics, m)
	}
	return metrics, err
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

			pollCount := int64(1)
			rand := rand.Float64()
			metrics := []structs.Metric{
				{MType: "counter", ID: "PollCount", Delta: &pollCount},
				{MType: "gauge", ID: "RandomValue", Value: &rand},
			}

			metrics = append(metrics, getRtmMetrics()...)

			memMetrics, err := getMemoryMetrics()
			if err != nil {
				fmt.Printf("ERROR failed to poll memory metrics: %s", err.Error())
			} else {
				metrics = append(metrics, memMetrics...)
			}

			cpuMetrics, err := getCPUMetrics()
			if err != nil {
				fmt.Printf("ERROR failed to poll cpu metrics: %s", err.Error())
			} else {
				metrics = append(metrics, cpuMetrics...)
			}

			err = storage.Agent.UpdateMetrics(metrics)
			if err != nil {
				log.Printf("ERROR poller failed to poll metrics: %s", err.Error())
			}
		}
	}
}

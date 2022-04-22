package reporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/zklevsha/go-musthave-devops/internal/agstore"
)

func send(url string) error {
	resp, err := http.Post(url, "text/plain", bytes.NewBufferString(""))
	if err != nil {
		return fmt.Errorf("an error occured %v", err)

	}
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
		}
		return fmt.Errorf("bad StatusCode: %s (URL: %s, Response Body: %s)",
			resp.Status, url, string(body))
	}
	defer resp.Body.Close()
	return nil
}

func reportGauges(serverSocket string) {
	agstore.Mutex.Lock()
	g := make(map[string]float64)
	for k, v := range agstore.Gauges {
		g[k] = v
	}
	agstore.Mutex.Unlock()
	for k, v := range g {
		err := send(fmt.Sprintf("http://%s/update/%s/%s/%f", serverSocket, "gauge", k, v))
		if err != nil {
			log.Printf("ERROR failed to send metic %s(%f): %s\n", k, v, err.Error())
		} else {
			log.Printf("INFO metric %s(%f) was sent\n", k, v)
		}

	}
}

func reportCounters(serverSocket string) {
	agstore.Mutex.Lock()
	c := make(map[string]int64)
	for k, v := range agstore.Counters {
		c[k] = v
	}
	agstore.Mutex.Unlock()

	for k, v := range c {
		err := send(fmt.Sprintf("http://%s/update/%s/%s/%d", serverSocket, "counter", k, v))
		if err != nil {
			log.Printf("ERROR failed to send metic %s(%d): %s\n", k, v, err.Error())
		} else {
			log.Printf("INFO metric %s(%d) was sent\n", k, v)
		}
	}
}

func Report(ctx context.Context, wg *sync.WaitGroup, serverSocket string, reportInterval time.Duration) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("INFO report received ctx.Done(), returning")
			return
		default:
			reportGauges(serverSocket)
			reportCounters(serverSocket)
			time.Sleep(reportInterval)
		}
	}
}

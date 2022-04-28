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

	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

func send(url string, body []byte) error {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
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
	url := fmt.Sprintf("http://%s/update/", serverSocket)
	for k, v := range storage.Agent.GetAllGauges() {
		body, err := serializer.EncodeBodyGauge(k, v)
		if err != nil {
			log.Printf("ERROR failed to convert metric %s (%f) to JSON: %s",
				k, v, err.Error())
			continue
		}
		err = send(url, body)
		if err != nil {
			log.Printf("ERROR failed to send metic %s(%f): %s\n", k, v, err.Error())
		} else {
			log.Printf("INFO metric %s(%f) was sent\n", k, v)
		}

	}
}

func reportCounters(serverSocket string) {
	url := fmt.Sprintf("http://%s/update/", serverSocket)
	for k, v := range storage.Agent.GetAllCounters() {
		body, err := serializer.EncodeBodyCounter(k, v)
		if err != nil {
			log.Printf("ERROR failed to convert metric %s (%d) to JSON: %s",
				k, v, err.Error())
			continue
		}
		err = send(url, body)
		if err != nil {
			log.Printf("ERROR failed to send metic %s(%d): %s\n", k, v, err.Error())
		} else {
			log.Printf("INFO metric %s(%d) was sent\n", k, v)
			storage.Agent.ResetCounter(k)
		}
	}
}

func Report(ctx context.Context, wg *sync.WaitGroup, serverSocket string, reportInterval time.Duration) {
	defer wg.Done()
	ticker := time.NewTicker(reportInterval)
	for {
		select {
		case <-ctx.Done():
			log.Println("INFO report received ctx.Done(), returning")
			return
		case <-ticker.C:
			reportGauges(serverSocket)
			reportCounters(serverSocket)
		}
	}
}

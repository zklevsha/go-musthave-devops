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

	"github.com/zklevsha/go-musthave-devops/internal/archive"
	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

func send(url string, body []byte) error {
	client := &http.Client{}

	compressed, err := archive.Compress(body)
	if err != nil {
		return fmt.Errorf("failed to compress request body: %s", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressed))
	if err != nil {
		return fmt.Errorf("failed to create http.NewRequest : %s", err.Error())
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")
	resp, err := client.Do(req)
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
			log.Printf("ERROR failed to convert metric %s (%s) to JSON: %s",
				k, body, err.Error())
			continue
		}
		err = send(url, body)
		if err != nil {
			log.Printf("ERROR failed to send metic %s(%s): %s\n", k, body, err.Error())
		} else {
			log.Printf("INFO metric %s(%s) was sent\n", k, body)
		}

	}
}

func reportCounters(serverSocket string) {
	url := fmt.Sprintf("http://%s/update/", serverSocket)
	for k, v := range storage.Agent.GetAllCounters() {
		body, err := serializer.EncodeBodyCounter(k, v)
		if err != nil {
			log.Printf("ERROR failed to convert metric %s (%s) to JSON: %s",
				k, body, err.Error())
			continue
		}
		err = send(url, body)
		if err != nil {
			log.Printf("ERROR failed to send metic %s(%s): %s\n", k, body, err.Error())
		} else {
			log.Printf("INFO metric %s(%s) was sent\n", k, body)
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

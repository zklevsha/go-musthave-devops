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
	"github.com/zklevsha/go-musthave-devops/internal/config"
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

func reportMetrics(serverSocket string, key string) {
	url := fmt.Sprintf("http://%s/updates/", serverSocket)
	metircs, _, err := storage.Agent.GetMetrics()
	if err != nil {
		log.Printf("ERROR failed to get metrics: %s", err.Error())
	}
	body, err := serializer.EncodeBodyMetrics(metircs, key)
	if err != nil {
		log.Printf("ERROR failed to encode metrics: %s", err.Error())
		return
	}
	err = send(url, body)
	if err != nil {
		log.Printf("ERROR failed to send metrics: %s", err.Error())
		return
	}
	log.Printf("INFO all metrics were sent")
	for _, m := range metircs {
		if m.MType == "counter" {
			err := storage.Agent.ResetCounter(m.ID)
			if err != nil {
				log.Printf("ERROR: failed to reset counter %s: %s", m.ID, err.Error())
			}
		}
	}
}

func Report(ctx context.Context, wg *sync.WaitGroup, conf config.AgentConfig) {
	defer wg.Done()
	ticker := time.NewTicker(conf.ReportInterval)
	for {
		select {
		case <-ctx.Done():
			log.Println("INFO report received ctx.Done(), returning")
			return
		case <-ticker.C:
			reportMetrics(conf.ServerAddress, conf.Key)
		}
	}
}

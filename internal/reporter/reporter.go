// Package reporter implements metric sending logic
package reporter

import (
	"bytes"
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/zklevsha/go-musthave-devops/internal/archive"
	"github.com/zklevsha/go-musthave-devops/internal/config"
	"github.com/zklevsha/go-musthave-devops/internal/rsaencrypt"
	"github.com/zklevsha/go-musthave-devops/internal/serializer"
	"github.com/zklevsha/go-musthave-devops/internal/storage"
)

func send(url string, body []byte, pubKey *rsa.PublicKey) error {
	client := &http.Client{}
	var b []byte
	var err error

	// Compress
	b, err = archive.Compress(body)
	if err != nil {
		return fmt.Errorf("failed to compress request body: %s", err.Error())
	}

	// Encrypt
	if pubKey != nil {
		b, err = rsaencrypt.Encrypt(pubKey, b, []byte(config.RsaLabel))
		if err != nil {
			return fmt.Errorf("ERROR failed to ecnrypt metrics: %s", err.Error())
		}
	}

	// Send
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create http.NewRequest : %s", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("an error occured %v", err)

	}

	// Check response
	if resp.StatusCode != 200 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
		}
		return fmt.Errorf("bad StatusCode: %s (URL: %s, Response Body: %s)",
			resp.Status, url, string(respBody))
	}
	defer resp.Body.Close()
	return nil
}

func reportMetrics(conf config.AgentConfig, pubKey *rsa.PublicKey) {
	url := fmt.Sprintf("http://%s/update/", conf.ServerAddress)
	metircs, err := storage.Agent.GetMetrics()
	if err != nil {
		log.Printf("ERROR failed to get metrics: %s", err.Error())
	}
	for _, m := range metircs {
		body, err := serializer.EncodyBodyMetric(m, conf.Key)
		if err != nil {
			log.Printf("ERROR failed to encode metrics: %s", err.Error())
			continue
		}
		err = send(url, body, pubKey)
		if err != nil {
			log.Printf("ERROR failed to send metric %s: %s", m.ID, err.Error())
			continue
		}
		log.Printf("INFO %s was sent", m.ID)
		if m.MType == "counter" {
			err := storage.Agent.ResetCounter(m.ID)
			if err != nil {
				log.Printf("ERROR: failed to reset counter %s: %s", m.ID, err.Error())
			}
		}
	}
}

func Report(ctx context.Context, wg *sync.WaitGroup, conf config.AgentConfig, pubKey *rsa.PublicKey) {
	defer wg.Done()
	ticker := time.NewTicker(conf.ReportInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("INFO report received ctx.Done(), returning")
			return
		case <-ticker.C:
			reportMetrics(conf, pubKey)
		}
	}
}

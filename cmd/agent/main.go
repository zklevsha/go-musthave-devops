package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zklevsha/go-musthave-devops/internal/poller"
	"github.com/zklevsha/go-musthave-devops/internal/reporter"
)

const pollIntervalDefault = time.Duration(2 * time.Second)
const reportIntervalDefault = time.Duration(10 * time.Second)
const serverAddressDefault = "127.0.0.1:8080"

var wg sync.WaitGroup

type agetnConfig struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	serverAddress  string
}

func getAgentConfig() agetnConfig {
	c := agetnConfig{
		pollInterval:   pollIntervalDefault,
		reportInterval: reportIntervalDefault,
		serverAddress:  serverAddressDefault}

	poll := os.Getenv("POLL_INTERVAL")
	if poll != "" {
		p, err := time.ParseDuration(poll)
		if err != nil {
			log.Printf("WARN main failed to parse env var POLL_INTERVAL=%s: %s. Default will be used (%v)",
				poll, err.Error(), c.pollInterval)
		}
		c.pollInterval = p
	}

	report := os.Getenv("REPORT_INTERVAL")
	if report != "" {
		r, err := time.ParseDuration(report)
		if err != nil {
			panic(err)
		}
		c.reportInterval = r
	}

	address := os.Getenv("ADDRESS")
	if address != "" {
		c.serverAddress = address
	}
	return c

}

func main() {
	agentConfig := getAgentConfig()
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	wg.Add(2)
	go poller.Poll(ctx, &wg, agentConfig.pollInterval)
	go reporter.Report(ctx, &wg, agentConfig.serverAddress, agentConfig.reportInterval)
	sig := <-c
	log.Printf("INFO main got a signal '%v', start shutting down...\n", sig)
	cancel()
	wg.Wait()
	log.Printf("INFO main Shutdown complete")
}

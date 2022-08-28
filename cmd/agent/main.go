package main

import (
	"context"
	"crypto/rsa"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/zklevsha/go-musthave-devops/internal/config"
	"github.com/zklevsha/go-musthave-devops/internal/poller"
	"github.com/zklevsha/go-musthave-devops/internal/reporter"
	"github.com/zklevsha/go-musthave-devops/internal/rsaencrypt"
)

var wg sync.WaitGroup

var buildVersion string = "N/A"
var buildDate string = "N/A"
var buildCommit string = "N/A"

func printStartupInfo() {
	log.Println("INFO main starting agent")
	log.Printf("INFO main Build version: %s", buildVersion)
	log.Printf("INFO main Build date: %s", buildDate)
	log.Printf("INFO main Build commit %s", buildCommit)
	log.Printf("DEBUG startup flags: %v", os.Args)
	log.Printf("DEBUG ENVs: %v", os.Environ())
}

func main() {
	printStartupInfo()
	agentConfig := config.GetAgentConfig()
	log.Printf("INFO main agent config: PollInterval: %v, ReportInterval: %v, ServerAddress: %s, PublicKeyPath: %s",
		agentConfig.PollInterval, agentConfig.ReportInterval, agentConfig.ServerAddress, agentConfig.PublicKeyPath)

	var pubKey *rsa.PublicKey
	var err error
	if agentConfig.PublicKeyPath != "" {
		pubKey, err = rsaencrypt.LoadPublicKey(agentConfig.PublicKeyPath)
		if err != nil {
			log.Fatalf("ERROR failed to load public key %s: %s", agentConfig.PublicKeyPath, err.Error())
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	// starting poller
	wg.Add(1)
	go poller.Poll(ctx, &wg, agentConfig.PollInterval)
	// starting reporter
	wg.Add(1)
	go reporter.Report(ctx, &wg, agentConfig, pubKey)

	// waiting for stop signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	sig := <-c
	log.Printf("INFO main got a signal '%v', start shutting down...\n", sig)
	cancel()
	wg.Wait()
	log.Printf("INFO main Shutdown complete")
}

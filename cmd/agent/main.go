package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/zklevsha/go-musthave-devops/internal/config"
	"github.com/zklevsha/go-musthave-devops/internal/poller"
	"github.com/zklevsha/go-musthave-devops/internal/reporter"
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
	log.Printf("INFO main agent config: PollInterval: %v, ReportInterval: %v, ServerAddress: %s",
		agentConfig.PollInterval, agentConfig.ReportInterval, agentConfig.ServerAddress)
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	wg.Add(2)
	go poller.Poll(ctx, &wg, agentConfig.PollInterval)
	go reporter.Report(ctx, &wg, agentConfig)
	sig := <-c
	log.Printf("INFO main got a signal '%v', start shutting down...\n", sig)
	cancel()
	wg.Wait()
	log.Printf("INFO main Shutdown complete")
}
